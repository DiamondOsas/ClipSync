package com.diamond.clipsync    

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import org.gioui.GioActivity // Gio i hate youuu

class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // 1. Set up the signal to target the Gio app
        val gioIntent = Intent(this, GioActivity::class.java)
        
        // 2. Send the signal to launch it!
        startActivity(gioIntent)
    }
}