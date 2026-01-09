package com.avtoplaneta.avtoplanetaapp.api;

import android.content.Context;
import android.content.SharedPreferences;

public class ApiConfig {
    private static final String PREFS_NAME = "api_config";
    private static final String KEY_BASE_URL = "base_url";
    private static final String DEFAULT_BASE_URL = "http://192.168.1.63:8080/";

    private final SharedPreferences prefs;

    public ApiConfig(Context context) {
        prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
    }

    public String getBaseUrl() {
        return prefs.getString(KEY_BASE_URL, DEFAULT_BASE_URL);
    }

    public void setBaseUrl(String baseUrl) {
        prefs.edit().putString(KEY_BASE_URL, baseUrl).apply();
    }

    public void resetToDefault() {
        setBaseUrl(DEFAULT_BASE_URL);
    }
}