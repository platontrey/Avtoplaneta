package com.avtoplaneta.avtoplanetaapp;

import android.app.Application;
import timber.log.Timber;

public class AvtoplanetaApplication extends Application {

    @Override
    public void onCreate() {
        super.onCreate();

        // Initialize Timber for better logging
        // Always initialize in debug mode (build config will be available at runtime)
        Timber.plant(new Timber.DebugTree());
    }
}