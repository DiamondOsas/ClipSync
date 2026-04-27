package com.diamond.clipsync

import org.junit.Assert.assertEquals
import org.junit.Test

class NetworkRepositoryTest {
    @Test
    fun testIpFormatting() {
        val ipInt = 16843009 // 1.1.1.1
        val formatted = com.diamond.clipsync.network.Formatter.formatIpAddress(ipInt)
        assertEquals("1.1.1.1", formatted)

        val ipInt2 = 2130706433 // 127.0.0.1
        val formatted2 = com.diamond.clipsync.network.Formatter.formatIpAddress(ipInt2)
        assertEquals("1.0.0.127", formatted2)
    }
}
