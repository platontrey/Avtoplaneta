package com.avtoplaneta.avtoplanetaapp.api;

import android.util.Log;

import com.avtoplaneta.avtoplanetaapp.BuildConfig;
import java.net.CookieManager;
import java.net.CookiePolicy;

import okhttp3.JavaNetCookieJar;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.logging.HttpLoggingInterceptor;
import retrofit2.Retrofit;
import retrofit2.converter.gson.GsonConverterFactory;

import java.util.concurrent.TimeUnit;

public class RetrofitClient {
    private static final String BASE_URL = BuildConfig.BASE_URL;
    private static Retrofit retrofit = null;
    private static String csrfToken = null;
    private static CookieManager cookieManager = null;

    public static Retrofit getClient() {
        if (retrofit == null) {
            // Логирование запросов
            HttpLoggingInterceptor logging = new HttpLoggingInterceptor();
            if (BuildConfig.DEBUG) {
                logging.setLevel(HttpLoggingInterceptor.Level.BODY);
            } else {
                logging.setLevel(HttpLoggingInterceptor.Level.NONE);
            }

            // Настройка cookie manager для поддержки сессий
            if (cookieManager == null) {
                cookieManager = new CookieManager();
                cookieManager.setCookiePolicy(CookiePolicy.ACCEPT_ALL);
            }

            OkHttpClient client = new OkHttpClient.Builder()
                    .cookieJar(new JavaNetCookieJar(cookieManager))
                    .addInterceptor(chain -> {
                        Request.Builder builder = chain.request().newBuilder();
                        // Добавляем CSRF токен, если есть
                        if (csrfToken != null && !csrfToken.isEmpty()) {
                            builder.addHeader("X-CSRF-Token", csrfToken);
                            Log.d("RetrofitClient", "CSRF token added: " + csrfToken);
                        } else {
                            Log.w("RetrofitClient", "CSRF token is null or empty");
                        }
                        Request request = builder.build();
                        Log.d("RetrofitClient", "Request URL: " + request.url() + ", Headers: " + request.headers());
                        return chain.proceed(request);
                    })
                    .addInterceptor(logging)
                    .connectTimeout(30, TimeUnit.SECONDS)
                    .readTimeout(30, TimeUnit.SECONDS)
                    .writeTimeout(30, TimeUnit.SECONDS)
                    .build();

            retrofit = new Retrofit.Builder()
                    .baseUrl(BASE_URL)
                    .client(client)
                    .addConverterFactory(GsonConverterFactory.create())
                    .build();
        }
        return retrofit;
    }

    public static ApiService getApiService() {
        return getClient().create(ApiService.class);
    }

    public static void setCsrfToken(String token) {
        csrfToken = token;
    }

    public static String getCsrfToken() {
        return csrfToken;
    }
}