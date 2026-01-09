package com.avtoplaneta.avtoplanetaapp.utils;

import android.content.Context;
import android.util.Log;
import android.widget.Toast;

public class NetworkErrorHandler {

    public static void handleNetworkError(Context context, String tag, String operation, Throwable t) {
        Log.e(tag, "Network error " + operation + ": " + t.getMessage(), t);

        String errorMessage = "Ошибка сети";
        if (t instanceof java.net.UnknownHostException) {
            errorMessage = "Не удается подключиться к серверу. Проверьте подключение к интернету и настройки сервера.";
        } else if (t instanceof java.net.ConnectException) {
            errorMessage = "Не удается подключиться к серверу. Возможно, сервер недоступен или IP адрес неправильный.";
        } else if (t instanceof java.net.SocketTimeoutException) {
            errorMessage = "Превышено время ожидания ответа сервера.";
        } else if (t instanceof java.io.IOException) {
            errorMessage = "Ошибка ввода-вывода: " + t.getMessage();
        }

        Toast.makeText(context, errorMessage, Toast.LENGTH_LONG).show();
    }

    public static void handleApiError(Context context, String tag, String operation, int statusCode, String responseBody) {
        Log.e(tag, "API error " + operation + ": HTTP " + statusCode + ", Response: " + responseBody);

        String errorMessage = "Ошибка сервера: " + statusCode;
        if (statusCode == 404) {
            errorMessage = "Ресурс не найден (404)";
        } else if (statusCode == 500) {
            errorMessage = "Внутренняя ошибка сервера (500)";
        } else if (statusCode == 403) {
            errorMessage = "Доступ запрещен (403)";
        } else if (statusCode == 401) {
            errorMessage = "Не авторизован (401)";
        }

        Toast.makeText(context, errorMessage, Toast.LENGTH_LONG).show();
    }
}