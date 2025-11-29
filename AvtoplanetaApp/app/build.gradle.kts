plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.hilt.android.plugin)
    id("kotlin-kapt")
}

android {
    namespace = "com.avtoplaneta.avtoplanetaapp"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.avtoplaneta.avtoplanetaapp"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    
    kotlin {
        jvmToolchain(17)
    }
    
    buildFeatures {
        viewBinding = true
    }
}

dependencies {
    implementation(libs.appcompat)
    implementation(libs.material)
    implementation(libs.activity)
    implementation(libs.constraintlayout)

    // Kotlin
    implementation("org.jetbrains.kotlin:kotlin-stdlib:2.2.0")

    // Hilt - Dependency Injection
    implementation(libs.hilt.android)
    kapt(libs.hilt.compiler)

    // ViewModel & LiveData
    implementation("androidx.lifecycle:lifecycle-viewmodel:2.8.4")
    implementation("androidx.lifecycle:lifecycle-livedata:2.8.4")
    implementation("androidx.lifecycle:lifecycle-viewmodel-ktx:2.8.4")
    implementation("androidx.lifecycle:lifecycle-livedata-ktx:2.8.4")

    // Room - Local Database
    implementation("androidx.room:room-runtime:2.6.1")
    implementation("androidx.room:room-ktx:2.6.1")
    kapt("androidx.room:room-compiler:2.6.1")

    // Retrofit для работы с API
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")

    // OkHttp для логирования
    implementation("com.squareup.okhttp3:logging-interceptor:4.11.0")

    // Gson для JSON
    implementation("com.google.code.gson:gson:2.10.1")

    // RecyclerView для списков
    implementation("androidx.recyclerview:recyclerview:1.3.2")
    implementation("androidx.cardview:cardview:1.0.0")

    // SwipeRefreshLayout
    implementation("androidx.swiperefreshlayout:swiperefreshlayout:1.1.0")

    // SharedPreferences для хранения сессии
    implementation("androidx.preference:preference:1.2.1")

    // Coil - Modern Image Loading (замена Glide)
    implementation("io.coil-kt:coil:2.6.0")
    
    // Glide - Image Loading
    implementation("com.github.bumptech.glide:glide:4.16.0")

    // Timber - Advanced Logging
    implementation("com.jakewharton.timber:timber:5.0.1")

    // LeakCanary - Memory Leak Detection
    debugImplementation("com.squareup.leakcanary:leakcanary-android:2.14")

    // Lottie - Animations
    implementation("com.airbnb.android:lottie:6.4.0")

    // PhotoView - Image Zoom
    implementation("io.getstream:photoview:1.0.3")

    // Image Compressor
    implementation("id.zelory:compressor:3.0.1")

    testImplementation(libs.junit)
    androidTestImplementation(libs.ext.junit)
    androidTestImplementation(libs.espresso.core)
}