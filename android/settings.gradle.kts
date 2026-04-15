pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
        flatDir {
            dirs("app/libs")   // ← add this
        }
    }
}

rootProject.name = "Clipsync"
include(":app")         // registers the "app" module