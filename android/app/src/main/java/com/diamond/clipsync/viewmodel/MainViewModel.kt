package com.diamond.clipsync.viewmodel

import android.app.Application
import android.content.Context
import android.content.Intent
import android.provider.Settings
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.diamond.clipsync.network.Device
import com.diamond.clipsync.network.NetworkRepository
import com.diamond.clipsync.service.AppState
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class MainViewModel(application: Application) : AndroidViewModel(application) {

    // Ideally we would use DI (like Hilt), but to keep it simple and modular:
    // If the accessibility service is not running, we can start a local NetworkRepository just for the UI
    // to see devices. But the Service is the source of truth for clipboard.

    private val networkRepository = NetworkRepository(application)

    val devices: StateFlow<List<Device>> = networkRepository.devices.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = emptyList()
    )

    val currentClipboard: StateFlow<String> = networkRepository.clipboardData.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = ""
    )

    val isServiceRunning: StateFlow<Boolean> = AppState.isServiceRunning

    init {
        // Start discovery for UI purposes if the service isn't doing it
        // (Though ideally they share the repository if they are in same process)
        viewModelScope.launch {
            networkRepository.startDiscovery()
        }
    }

    fun openAccessibilitySettings(context: Context) {
        val intent = Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS)
        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        context.startActivity(intent)
    }

    override fun onCleared() {
        super.onCleared()
        // We only stop the UI's network repo. The Service has its own.
        networkRepository.stop()
    }
}
