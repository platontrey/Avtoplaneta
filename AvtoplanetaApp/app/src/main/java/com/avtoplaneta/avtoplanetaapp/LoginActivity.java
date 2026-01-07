package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Bundle;
import android.text.TextUtils;
import android.view.View;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;

import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.ActivityLoginBinding;
import com.avtoplaneta.avtoplanetaapp.models.CsrfResponse;
import com.avtoplaneta.avtoplanetaapp.models.LoginRequest;
import com.avtoplaneta.avtoplanetaapp.models.LoginResponse;
import com.google.gson.Gson;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class LoginActivity extends AppCompatActivity {

    private ActivityLoginBinding binding;
    private SharedPreferences sharedPreferences;
    private static final String PREFS_NAME = "AvtoplanetaPrefs";
    private static final String KEY_USER_DATA = "user_data";
    private static final String KEY_IS_LOGGED_IN = "is_logged_in";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        binding = ActivityLoginBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        sharedPreferences = getSharedPreferences(PREFS_NAME, MODE_PRIVATE);

        // Проверка, если пользователь уже вошел
        if (isUserLoggedIn()) {
            navigateToInventory();
            return;
        }

        setupListeners();
    }

    private void setupListeners() {
        binding.btnLogin.setOnClickListener(v -> attemptLogin());
    }

    private void attemptLogin() {
        // Скрыть предыдущую ошибку
        binding.tvError.setVisibility(View.GONE);

        // Получить данные из полей
        String email = binding.etEmail.getText().toString().trim();
        String password = binding.etPassword.getText().toString().trim();

        // Валидация
        if (TextUtils.isEmpty(email)) {
            binding.tilEmail.setError("Введите email или имя пользователя");
            return;
        }

        if (TextUtils.isEmpty(password)) {
            binding.tilPassword.setError("Введите пароль");
            return;
        }

        // Очистить ошибки
        binding.tilEmail.setError(null);
        binding.tilPassword.setError(null);

        // Показать индикатор загрузки
        setLoading(true);

        // Создать запрос
        LoginRequest loginRequest = new LoginRequest(email, password);

        // Выполнить API запрос
        RetrofitClient.getApiService().login(loginRequest).enqueue(new Callback<LoginResponse>() {
            @Override
            public void onResponse(Call<LoginResponse> call, Response<LoginResponse> response) {
                setLoading(false);

                if (response.isSuccessful() && response.body() != null) {
                    LoginResponse loginResponse = response.body();

                    if (loginResponse.getUser() != null) {
                        // Сохранить данные пользователя
                        saveUserData(loginResponse);

                        // Получить CSRF токен
                        fetchCsrfToken();
                    } else {
                        showError("Ошибка авторизации");
                    }
                } else {
                    showError("Неверный email или пароль");
                }
            }

            @Override
            public void onFailure(Call<LoginResponse> call, Throwable t) {
                setLoading(false);
                showError("Ошибка соединения: " + t.getMessage());
            }
        });
    }

    private void saveUserData(LoginResponse loginResponse) {
        SharedPreferences.Editor editor = sharedPreferences.edit();
        Gson gson = new Gson();
        String userJson = gson.toJson(loginResponse.getUser());
        editor.putString(KEY_USER_DATA, userJson);
        editor.putBoolean(KEY_IS_LOGGED_IN, true);
        editor.apply();
    }

    private boolean isUserLoggedIn() {
        return sharedPreferences.getBoolean(KEY_IS_LOGGED_IN, false);
    }

    private void fetchCsrfToken() {
        RetrofitClient.getApiService().getCsrfToken().enqueue(new Callback<CsrfResponse>() {
            @Override
            public void onResponse(Call<CsrfResponse> call, Response<CsrfResponse> response) {
                if (response.isSuccessful() && response.body() != null) {
                    String csrfToken = response.body().getCsrfToken();
                    RetrofitClient.setCsrfToken(csrfToken);
                    // Перейти к Inventory
                    navigateToInventory();
                } else {
                    // Даже если не удалось получить токен, продолжаем
                    navigateToInventory();
                }
            }

            @Override
            public void onFailure(Call<CsrfResponse> call, Throwable t) {
                // Даже если не удалось получить токен, продолжаем
                navigateToInventory();
            }
        });
    }

    private void navigateToInventory() {
        Intent intent = new Intent(LoginActivity.this, MainActivity.class);
        intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TASK);
        startActivity(intent);
        finish();
    }

    private void setLoading(boolean isLoading) {
        binding.progressBar.setVisibility(isLoading ? View.VISIBLE : View.GONE);
        binding.btnLogin.setEnabled(!isLoading);
        binding.etEmail.setEnabled(!isLoading);
        binding.etPassword.setEnabled(!isLoading);
    }

    private void showError(String message) {
        binding.tvError.setText(message);
        binding.tvError.setVisibility(View.VISIBLE);
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        binding = null;
    }
}