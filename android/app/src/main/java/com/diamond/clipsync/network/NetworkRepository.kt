package com.diamond.clipsync.network

import android.content.Context
import android.net.wifi.WifiManager
import android.util.Log
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress
import javax.jmdns.JmDNS
import javax.jmdns.ServiceEvent
import javax.jmdns.ServiceInfo
import javax.jmdns.ServiceListener

data class Device(val name: String, val ip: String)

class NetworkRepository(private val context: Context) {

    private val _devices = MutableStateFlow<List<Device>>(emptyList())
    val devices: StateFlow<List<Device>> = _devices.asStateFlow()

    private val _clipboardData = MutableStateFlow<String>("")
    val clipboardData: StateFlow<String> = _clipboardData.asStateFlow()

    private var jmdns: JmDNS? = null
    private var udpSocket: DatagramSocket? = null
    private var isListening = false

    private val SERVICE_TYPE = "_clipsync._tcp.local."
    private val PORT = 9999
    private val TAG = "NetworkRepository"

    private val serviceListener = object : ServiceListener {
        override fun serviceAdded(event: ServiceEvent) {
            Log.d(TAG, "Service added: \${event.info}")
            jmdns?.requestServiceInfo(event.type, event.name, 1)
        }

        override fun serviceRemoved(event: ServiceEvent) {
            Log.d(TAG, "Service removed: \${event.name}")
            _devices.value = _devices.value.filter { it.name != event.name }
        }

        override fun serviceResolved(event: ServiceEvent) {
            Log.d(TAG, "Service resolved: \${event.info}")
            val addresses = event.info.hostAddresses
            if (addresses.isNotEmpty()) {
                val ip = addresses[0]
                val newDevice = Device(event.name, ip)
                if (!_devices.value.contains(newDevice)) {
                    val currentList = _devices.value.toMutableList()
                    currentList.add(newDevice)
                    _devices.value = currentList

                    // Send an initial handshake/ready message similar to Go
                    sendToUdp("---ClipSync---", ip)
                }
            }
        }
    }

    suspend fun startDiscovery() = withContext(Dispatchers.IO) {
        try {
            val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as WifiManager
            val multicastLock = wifiManager.createMulticastLock("clipsync_multicast")
            multicastLock.setReferenceCounted(true)
            multicastLock.acquire()

            val ipAddress = Formatter.formatIpAddress(wifiManager.connectionInfo.ipAddress)
            val inetAddress = InetAddress.getByName(ipAddress)

            jmdns = JmDNS.create(inetAddress, android.os.Build.MODEL)
            jmdns?.addServiceListener(SERVICE_TYPE, serviceListener)

            // Register self
            val serviceInfo = ServiceInfo.create(
                SERVICE_TYPE,
                android.os.Build.MODEL,
                PORT,
                0,
                0,
                "ClipSync Android"
            )
            jmdns?.registerService(serviceInfo)

            startListeningUdp()

        } catch (e: Exception) {
            Log.e(TAG, "Error starting discovery", e)
        }
    }

    private fun startListeningUdp() {
        if (isListening) return
        isListening = true

        Thread {
            try {
                udpSocket = DatagramSocket(PORT)
                udpSocket?.broadcast = true
                val buffer = ByteArray(1024)

                while (isListening) {
                    val packet = DatagramPacket(buffer, buffer.size)
                    udpSocket?.receive(packet)

                    val receivedData = String(packet.data, 0, packet.length)
                    if (receivedData != "---ClipSync---") {
                        Log.d(TAG, "Received clipboard: \$receivedData")
                        _clipboardData.value = receivedData
                    } else {
                        // It's a handshake, we can add the IP if not already there
                        val senderIp = packet.address.hostAddress
                        Log.d(TAG, "Received handshake from: \$senderIp")
                        if (senderIp != null) {
                            val existingDevice = _devices.value.find { it.ip == senderIp }
                            if (existingDevice == null) {
                                val currentList = _devices.value.toMutableList()
                                currentList.add(Device("Unknown (\${senderIp})", senderIp))
                                _devices.value = currentList
                            }
                        }
                    }
                }
            } catch (e: Exception) {
                if (isListening) {
                    Log.e(TAG, "Error listening to UDP", e)
                }
            }
        }.start()
    }

    fun sendClipboard(text: String) {
        val currentDevices = _devices.value
        Thread {
            currentDevices.forEach { device ->
                sendToUdp(text, device.ip)
            }
        }.start()
    }

    private fun sendToUdp(text: String, ip: String) {
        try {
            val address = InetAddress.getByName(ip)
            val data = text.toByteArray()
            val packet = DatagramPacket(data, data.size, address, PORT)

            val tempSocket = DatagramSocket()
            tempSocket.send(packet)
            tempSocket.close()
        } catch (e: Exception) {
            Log.e(TAG, "Error sending to UDP \$ip", e)
        }
    }

    fun stop() {
        isListening = false
        udpSocket?.close()
        jmdns?.removeServiceListener(SERVICE_TYPE, serviceListener)
        jmdns?.unregisterAllServices()
        jmdns?.close()
    }
}
