package com.avtoplaneta.avtoplanetaapp;

import static com.avtoplaneta.avtoplanetaapp.R.*;

import android.annotation.SuppressLint;
import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Bundle;
import android.text.Editable;
import android.text.TextWatcher;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.view.WindowManager;
import android.widget.EditText;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.view.WindowCompat;
import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentTransaction;
import androidx.recyclerview.widget.GridLayoutManager;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.google.android.material.bottomnavigation.BottomNavigationView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.ActivityMainBinding;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.avtoplaneta.avtoplanetaapp.models.User;
import com.google.gson.Gson;

import java.util.ArrayList;
import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class MainActivity extends AppCompatActivity {

    private ActivityMainBinding binding;
    private SharedPreferences sharedPreferences;
    private static final String PREFS_NAME = "AvtoplanetaPrefs";
    private static final String KEY_IS_LOGGED_IN = "is_logged_in";
    private static final String KEY_USER_DATA = "user_data";

    private List<InventoryItem> allParts;
    private PartsAdapter partsAdapter;

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

        // Инициализация данных
        allParts = new ArrayList<>();

        // Настройка адаптеров
        setupAdapters();

        // Настройка UI
        setupUI();

        // Настройка нижнего меню навигации
        setupBottomNavigation();

        // Настройка меню (бургер)
        setupMenu();

        // Загрузка данных
        loadPartsData();
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        Log.d("MainActivity", "onActivityResult called: requestCode=" + requestCode + ", resultCode=" + resultCode);

        if (requestCode == 100 && resultCode == RESULT_OK && data != null) {
            String action = data.getStringExtra("action");
            Log.d("MainActivity", "Processing action: " + action);

            if ("create".equals(action)) {
                // Add new part to the list with real ID from server
                InventoryItem newItem = new InventoryItem();
                newItem.setId(data.getIntExtra("part_id", allParts.size() + 1));
                newItem.setName(data.getStringExtra("part_name"));
                newItem.setBrand(data.getStringExtra("part_brand"));
                newItem.setModel(data.getStringExtra("part_model"));
                newItem.setCategory(data.getStringExtra("part_category"));
                newItem.setLocation(data.getStringExtra("part_location"));
                newItem.setPrice(data.getDoubleExtra("part_price", 0));
                newItem.setQuantity(data.getIntExtra("part_quantity", 0));
                newItem.setDescription(data.getStringExtra("part_description"));
                newItem.setStatus(true);
                newItem.setPhoto(""); // No photo for demo

                allParts.add(newItem);
                updateUI();
                Log.d("MainActivity", "Added new part: " + newItem.getName());

                // Refresh data from server to ensure consistency
                loadPartsData();

            } else if ("update".equals(action)) {
                // Update existing part
                int partId = data.getIntExtra("part_id", -1);
                Log.d("MainActivity", "Updating part with id: " + partId);
                boolean found = false;
                for (InventoryItem item : allParts) {
                    if (item.getId() == partId) {
                        item.setName(data.getStringExtra("part_name"));
                        item.setBrand(data.getStringExtra("part_brand"));
                        item.setModel(data.getStringExtra("part_model"));
                        item.setCategory(data.getStringExtra("part_category"));
                        item.setLocation(data.getStringExtra("part_location"));
                        item.setPrice(data.getDoubleExtra("part_price", 0));
                        item.setQuantity(data.getIntExtra("part_quantity", 0));
                        item.setDescription(data.getStringExtra("part_description"));
                        found = true;
                        Log.d("MainActivity", "Updated part: " + item.getName());
                        break;
                    }
                }
                if (found) {
                    updateUI();
                    Log.d("MainActivity", "UI updated after part update");

                    // Refresh data from server to ensure consistency
                    loadPartsData();
                } else {
                    Log.e("MainActivity", "Part with id " + partId + " not found for update");
                }
            }
        } else {
            Log.d("MainActivity", "onActivityResult ignored: requestCode=" + requestCode + ", resultCode=" + resultCode);
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

    private void setupAdapters() {
        partsAdapter = new PartsAdapter(allParts, null);
        binding.partsRecyclerView.setLayoutManager(new LinearLayoutManager(this));
        binding.partsRecyclerView.setAdapter(partsAdapter);
    }

    private void setupUI() {
        // Настройка поиска
        binding.searchEditText.addTextChangedListener(new TextWatcher() {
            @Override
            public void beforeTextChanged(CharSequence s, int start, int count, int after) {}

            @Override
            public void onTextChanged(CharSequence s, int start, int before, int count) {
                filterParts(s.toString());
            }

            @Override
            public void afterTextChanged(Editable s) {}
        });
    }

    private void loadPartsData() {
        Log.d("MainActivity", "Loading parts data");
        ApiService apiService = RetrofitClient.getApiService();
        Call<List<InventoryItem>> call = apiService.getInventory();

        call.enqueue(new Callback<List<InventoryItem>>() {
            @Override
            public void onResponse(Call<List<InventoryItem>> call, Response<List<InventoryItem>> response) {
                if (response.isSuccessful() && response.body() != null) {
                    allParts.clear();
                    allParts.addAll(response.body());

                    // Создаем фиктивные данные для демонстрации
                    createSampleData();

                    updateUI();
                    Log.d("MainActivity", "Parts loaded: " + allParts.size());
                } else {
                    Log.e("MainActivity", "Error loading parts: " + response.code());
                }
            }

            @Override
            public void onFailure(Call<List<InventoryItem>> call, Throwable t) {
                Log.e("MainActivity", "Network error: " + t.getMessage());
                // Создаем фиктивные данные даже при ошибке сети
                createSampleData();
                updateUI();
            }
        });
    }

    private void createSampleData() {
        // Добавляем фиктивные данные для демонстрации
        if (allParts.isEmpty()) {
            InventoryItem item1 = new InventoryItem();
            item1.setId(1);
            item1.setName("Двигатель BMW N54");
            item1.setCategory("Двигатель");
            item1.setBrand("BMW");
            item1.setModel("N54");
            item1.setLocation("Склад A");
            item1.setDescription("Мощный турбодвигатель");
            item1.setPrice(150000.0);
            item1.setQuantity(5);
            item1.setStatus(true);
            item1.setPhoto("/uploads/engine1.jpg");
            allParts.add(item1);

            InventoryItem item2 = new InventoryItem();
            item2.setId(2);
            item2.setName("Трансмиссия Mercedes W124");
            item2.setCategory("Трансмиссия");
            item2.setBrand("Mercedes");
            item2.setModel("W124");
            item2.setLocation("Склад B");
            item2.setDescription("Автоматическая коробка передач");
            item2.setPrice(75000.0);
            item2.setQuantity(2);
            item2.setStatus(true);
            item2.setPhoto("/uploads/transmission1.jpg");
            allParts.add(item2);

            InventoryItem item3 = new InventoryItem();
            item3.setId(3);
            item3.setName("Тормозная система Audi A6");
            item3.setCategory("Тормоза");
            item3.setBrand("Audi");
            item3.setModel("A6");
            item3.setLocation("Склад A");
            item3.setDescription("Комплект тормозных дисков и колодок");
            item3.setPrice(25000.0);
            item3.setQuantity(8);
            item3.setStatus(true);
            item3.setPhoto("/uploads/brakes1.jpg");
            allParts.add(item3);

            InventoryItem item4 = new InventoryItem();
            item4.setId(4);
            item4.setName("Подвеска Volkswagen Golf");
            item4.setCategory("Подвеска");
            item4.setBrand("Volkswagen");
            item4.setModel("Golf");
            item4.setLocation("Склад C");
            item4.setDescription("Полный комплект подвески");
            item4.setPrice(35000.0);
            item4.setQuantity(3);
            item4.setStatus(true);
            item4.setPhoto("/uploads/suspension1.jpg");
            allParts.add(item4);

            InventoryItem item5 = new InventoryItem();
            item5.setId(5);
            item5.setName("Аккумулятор Bosch");
            item5.setCategory("Электрика");
            item5.setBrand("Bosch");
            item5.setModel("S4");
            item5.setLocation("Склад A");
            item5.setDescription("Высококачественный аккумулятор");
            item5.setPrice(8000.0);
            item5.setQuantity(15);
            item5.setStatus(true);
            item5.setPhoto("/uploads/battery1.jpg");
            allParts.add(item5);

            InventoryItem item6 = new InventoryItem();
            item6.setId(6);
            item6.setName("Фары LED H7");
            item6.setCategory("Оптика");
            item6.setBrand("Philips");
            item6.setModel("LED H7");
            item6.setLocation("Склад B");
            item6.setDescription("Комплект LED фар");
            item6.setPrice(12000.0);
            item6.setQuantity(10);
            item6.setStatus(true);
            item6.setPhoto("/uploads/headlights1.jpg");
            allParts.add(item6);
        }

    }

    private void updateUI() {
        partsAdapter.updateParts(allParts);

        // Обновляем статистику
        int totalParts = allParts.size();
        int availableParts = (int) allParts.stream().filter(InventoryItem::isStatus).count();

        binding.tvTotalParts.setText(String.valueOf(totalParts));
        binding.tvAvailableParts.setText(String.valueOf(availableParts));
    }

    private void filterParts(String query) {
        if (query == null || query.trim().isEmpty()) {
            partsAdapter.updateParts(allParts);
        } else {
            String lowerQuery = query.toLowerCase().trim();
            List<InventoryItem> filtered = new ArrayList<>();
            for (InventoryItem item : allParts) {
                if (item.getName().toLowerCase().contains(lowerQuery) ||
                    item.getBrand().toLowerCase().contains(lowerQuery) ||
                    item.getModel().toLowerCase().contains(lowerQuery) ||
                    item.getCategory().toLowerCase().contains(lowerQuery)) {
                    filtered.add(item);
                }
            }
            partsAdapter.updateParts(filtered);
        }
    }

    private void setupBottomNavigation() {
        // Bottom navigation теперь простой LinearLayout с ImageView элементами
        Log.d("MainActivity", "Bottom navigation setup - using ImageView layout");

        // Add click listeners for bottom navigation items
        binding.ivNavHome.setOnClickListener(v -> {
            Log.d("MainActivity", "Home clicked - already on home");
            // Уже на главной странице - ничего не делаем
        });

        binding.ivNavAdd.setOnClickListener(v -> {
            Log.d("MainActivity", "Add clicked - opening add part screen");
            Intent intent = new Intent(MainActivity.this, AddPartActivity.class);
            startActivityForResult(intent, 100);
        });

        binding.ivNavCart.setOnClickListener(v -> {
            Log.d("MainActivity", "Cart clicked - opening orders screen");
            Intent intent = new Intent(MainActivity.this, OrdersActivity.class);
            startActivity(intent);
        });
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