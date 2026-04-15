plugins { 
    // AGP 8.7.0 or higher is required for Gradle 9.x compatibility.
    // 'apply false' means the plugin is declared here to set the version, 
    // but not applied to the root project itself (only to subprojects like :app).
    id("com.android.application") version "8.7.0" apply false
    id("org.jetbrains.kotlin.android") version "2.0.0" apply false
}