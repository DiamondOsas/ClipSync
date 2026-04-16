plugins {
    id("com.android.application") 
    id("org.jetbrains.kotlin.android") 
}

android {
    namespace = "com.diamond.clipsync"
    compileSdk = 33

    defaultConfig {
        applicationId = "com.diamond.clipsync"
        minSdk = 24
        targetSdk = 33  
        versionCode = 1
        versionName = "1.0"
    }
    
    buildTypes {
        release {
            isMinifyEnabled = false    // set true to shrink/obfuscate code
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }
}

dependencies {
    implementation(fileTree(mapOf("dir" to "libs", "include" to listOf("*.aar", "*.jar"))))
}