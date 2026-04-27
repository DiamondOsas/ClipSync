package com.diamond.clipsync.service

import android.accessibilityservice.AccessibilityService
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.accessibility.AccessibilityEvent
import com.diamond.clipsync.network.NetworkRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class ClipSyncAccessibilityService : AccessibilityService() {

    private val TAG = "ClipSyncAccessibility"
    private var lastClipboardText: String = ""
    private var networkRepository: NetworkRepository? = null
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var clipboardManager: ClipboardManager? = null

    // Polling because AccessibilityEvent doesn't have a specific event for clipboard changes
    // globally without capturing text field changes, and clipboard manager listeners
    // don't work reliably in background on Android 10+
    private val handler = Handler(Looper.getMainLooper())
    private val clipboardPoller = object : Runnable {
        override fun run() {
            checkClipboard()
            handler.postDelayed(this, 2000) // Poll every 2 seconds
        }
    }

    override fun onServiceConnected() {
        super.onServiceConnected()
        com.diamond.clipsync.service.AppState.setServiceRunning(true)
        Log.d(TAG, "Accessibility Service Connected")

        clipboardManager = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
        networkRepository = NetworkRepository(this)

        scope.launch {
            networkRepository?.startDiscovery()

            networkRepository?.clipboardData?.collect { newText ->
                if (newText.isNotEmpty() && newText != lastClipboardText) {
                    setLocalClipboard(newText)
                }
            }
        }

        handler.post(clipboardPoller)
    }

    private fun checkClipboard() {
        try {
            if (clipboardManager?.hasPrimaryClip() == true) {
                val clipData = clipboardManager?.primaryClip
                if (clipData != null && clipData.itemCount > 0) {
                    val text = clipData.getItemAt(0).text?.toString()
                    if (text != null && text != lastClipboardText && text.isNotEmpty()) {
                        Log.d(TAG, "New clipboard detected: \$text")
                        lastClipboardText = text
                        networkRepository?.sendClipboard(text)
                    }
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error checking clipboard", e)
        }
    }

    private fun setLocalClipboard(text: String) {
        Handler(Looper.getMainLooper()).post {
            try {
                lastClipboardText = text
                val clip = ClipData.newPlainText("ClipSync", text)
                clipboardManager?.setPrimaryClip(clip)
                Log.d(TAG, "Local clipboard updated from network")
            } catch (e: Exception) {
                Log.e(TAG, "Error setting local clipboard", e)
            }
        }
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        // We can capture text selection or copy events here if polling isn't enough
        if (event?.eventType == AccessibilityEvent.TYPE_VIEW_TEXT_CHANGED ||
            event?.eventType == AccessibilityEvent.TYPE_VIEW_TEXT_SELECTION_CHANGED) {
            // Can be used to enhance detection if polling misses something
            // But frequent checking here might be too intensive, so polling is safer
        }
    }

    override fun onInterrupt() {
        Log.d(TAG, "Accessibility Service Interrupted")
    }

    override fun onUnbind(intent: Intent?): Boolean {
        handler.removeCallbacks(clipboardPoller)
        networkRepository?.stop()
        com.diamond.clipsync.service.AppState.setServiceRunning(false)
        return super.onUnbind(intent)
    }
}
