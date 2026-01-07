package com.avtoplaneta.avtoplanetaapp;

import static com.avtoplaneta.avtoplanetaapp.R.*;

import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Bundle;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.view.WindowCompat;
import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentManager;
import androidx.fragment.app.FragmentTransaction;

import com.avtoplaneta.avtoplanetaapp.databinding.ActivityMainBinding;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class MainActivity extends AppCompatActivity {

    private ActivityMainBinding binding;
    private SharedPreferences sharedPreferences;
    private static final String PREFS_NAME = "AvtoplanetaPrefs";
    private static final String KEY_IS_LOGGED_IN = "is_logged_in";
    private static final String KEY_USER_DATA = "user_data";

    private FragmentManager fragmentManager;
    private HomeFragment homeFragment;
    private OrdersFragment ordersFragment;
    private StatisticsFragment statisticsFragment;
    private AddPartFragment addPartFragment;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("MainActivity", "=== MAIN ACTIVITY STARTED ===");

        binding = ActivityMainBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        // Настройка прозрачного статус бара
        setupTransparentStatusBar();

        sharedPreferences = getSharedPreferences(PREFS_NAME, MODE_PRIVATE);

        // Проверить статус авторизации
        boolean isLoggedIn = sharedPreferences.getBoolean(KEY_IS_LOGGED_IN, false);
        Log.d("MainActivity", "Login status: " + isLoggedIn);

        if (!isLoggedIn) {
            // Пользователь не авторизован - перейти к Login
            Log.d("MainActivity", "User not logged in, redirecting to login");
            Intent intent = new Intent(MainActivity.this, LoginActivity.class);
            intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TASK);
            startActivity(intent);
            finish();
            return;
        }

        // Инициализация Fragments
        fragmentManager = getSupportFragmentManager();
        homeFragment = new HomeFragment();
        ordersFragment = new OrdersFragment();
        statisticsFragment = new StatisticsFragment();
        addPartFragment = new AddPartFragment();

        // Показать HomeFragment по умолчанию
        showFragment(homeFragment);

        // Настройка нижнего меню навигации
        setupBottomNavigation();

        // Настройка меню (бургер)
        setupMenu();
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        Log.d("MainActivity", "onActivityResult called: requestCode=" + requestCode + ", resultCode=" + resultCode);

        // Forward to HomeFragment if it's visible
        if (homeFragment != null && homeFragment.isVisible()) {
            homeFragment.onActivityResult(requestCode, resultCode, data);
        }
    }

    private void setupTransparentStatusBar() {
        WindowCompat.setDecorFitsSystemWindows(getWindow(), false);
        getWindow().setStatusBarColor(android.graphics.Color.TRANSPARENT);
        getWindow().getDecorView().setSystemUiVisibility(
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_STABLE |
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
        );
    }

    private void setupBottomNavigation() {
        // Bottom navigation теперь простой LinearLayout с ImageView элементами
        Log.d("MainActivity", "Bottom navigation setup - using ImageView layout");

        // Add click listeners for bottom navigation items
        binding.ivNavHome.setOnClickListener(v -> {
            Log.d("MainActivity", "Home clicked");
            showFragment(homeFragment);
        });

        binding.ivNavAdd.setOnClickListener(v -> {
            Log.d("MainActivity", "Add clicked - opening add part screen");
            Intent intent = new Intent(MainActivity.this, AddPartActivity.class);
            startActivityForResult(intent, 100);
        });

        binding.ivNavCart.setOnClickListener(v -> {
            Log.d("MainActivity", "Cart clicked - showing orders fragment");
            showFragment(ordersFragment);
        });

        binding.ivNavStats.setOnClickListener(v -> {
            Log.d("MainActivity", "Stats clicked - showing statistics fragment");
            showFragment(statisticsFragment);
        });
    }

    public void showFragment(Fragment fragment) {
        FragmentTransaction transaction = fragmentManager.beginTransaction();
        transaction.replace(R.id.fragmentContainer, fragment);
        transaction.commit();
    }

    private void setupMenu() {
        binding.menuIcon.setOnClickListener(v -> {
            Log.d("MainActivity", "Menu icon clicked - showing fullscreen menu");

            // Создаем полноэкранное меню с размытием
            showFullscreenMenu();
        });
    }

    private void showFullscreenMenu() {
        // Создаем View для полноэкранного меню
        View menuView = getLayoutInflater().inflate(R.layout.dialog_fullscreen_menu, null);

        // Находим элементы меню
        TextView tvLogout = menuView.findViewById(R.id.tvLogout);

        // Создаем диалог
        androidx.appcompat.app.AlertDialog dialog = new androidx.appcompat.app.AlertDialog.Builder(this, R.style.FullScreenDialog)
                .setView(menuView)
                .create();

        // Настраиваем обработчики
        tvLogout.setOnClickListener(v -> {
            Log.d("MainActivity", "Logout selected from fullscreen menu");
            dialog.dismiss();

            // Показываем диалог подтверждения выхода
            new androidx.appcompat.app.AlertDialog.Builder(this)
                    .setTitle("Выход")
                    .setMessage("Вы действительно хотите выйти из аккаунта?")
                    .setPositiveButton("Выйти", (confirmDialog, which) -> {
                        Log.d("MainActivity", "User confirmed logout");
                        logout();
                    })
                    .setNegativeButton("Отмена", null)
                    .show();
        });

        // Закрываем меню при нажатии вне его
        menuView.setOnClickListener(v -> dialog.dismiss());

        dialog.show();
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        getMenuInflater().inflate(R.menu.menu_inventory, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        if (item.getItemId() == R.id.action_logout) {
            logout();
            return true;
        }
        return super.onOptionsItemSelected(item);
    }

    private void logout() {
        // Очистить данные пользователя
        SharedPreferences.Editor editor = sharedPreferences.edit();
        editor.remove(KEY_USER_DATA);
        editor.putBoolean(KEY_IS_LOGGED_IN, false);
        editor.apply();

        // Вернуться к экрану входа
        Intent intent = new Intent(MainActivity.this, LoginActivity.class);
        intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TASK);
        startActivity(intent);
        finish();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        binding = null;
    }
}

