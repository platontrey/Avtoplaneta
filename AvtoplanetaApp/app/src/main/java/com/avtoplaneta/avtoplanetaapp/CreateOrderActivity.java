package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.recyclerview.widget.LinearLayoutManager;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.ActivityCreateOrderBinding;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.avtoplaneta.avtoplanetaapp.models.Order;
import com.avtoplaneta.avtoplanetaapp.models.OrderItem;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class CreateOrderActivity extends AppCompatActivity implements OrderPartsAdapter.OnSelectionChangedListener {

    private ActivityCreateOrderBinding binding;
    private OrderPartsAdapter adapter;
    private List<InventoryItem> partsList = new ArrayList<>();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("CreateOrderActivity", "=== CREATE ORDER ACTIVITY STARTED ===");

        binding = ActivityCreateOrderBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        setupUI();
        setupBottomNavigation();
        loadParts();
    }

    private void setupUI() {
        // Back button
        binding.ivBack.setOnClickListener(v -> {
            Log.d("CreateOrderActivity", "Back button clicked");
            finish();
        });

        // Setup RecyclerView
        binding.rvParts.setLayoutManager(new LinearLayoutManager(this));
        adapter = new OrderPartsAdapter(partsList, this);
        binding.rvParts.setAdapter(adapter);

        // Create order button
        binding.btnCreateOrder.setOnClickListener(v -> {
            Log.d("CreateOrderActivity", "Create order button clicked");
            createOrder();
        });
    }

    private void loadParts() {
        Log.d("CreateOrderActivity", "Loading parts from API");

        ApiService apiService = RetrofitClient.getApiService();
        Call<List<InventoryItem>> call = apiService.getInventory();

        call.enqueue(new Callback<List<InventoryItem>>() {
            @Override
            public void onResponse(Call<List<InventoryItem>> call, Response<List<InventoryItem>> response) {
                if (response.isSuccessful() && response.body() != null) {
                    partsList.clear();
                    partsList.addAll(response.body());
                    adapter.notifyDataSetChanged();
                    Log.d("CreateOrderActivity", "Loaded " + partsList.size() + " parts");
                } else {
                    Log.e("CreateOrderActivity", "Error loading parts: " + response.code());
                    Toast.makeText(CreateOrderActivity.this, "Ошибка загрузки запчастей", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<List<InventoryItem>> call, Throwable t) {
                Log.e("CreateOrderActivity", "Network error: " + t.getMessage());
                Toast.makeText(CreateOrderActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void setupBottomNavigation() {
        // Bottom navigation
        binding.ivNavHome.setOnClickListener(v -> {
            Log.d("CreateOrderActivity", "Home navigation clicked");
            Intent intent = new Intent(CreateOrderActivity.this, MainActivity.class);
            intent.setFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP);
            startActivity(intent);
            finish();
        });

        binding.ivNavAdd.setOnClickListener(v -> {
            Log.d("CreateOrderActivity", "Add navigation clicked");
            Intent intent = new Intent(CreateOrderActivity.this, AddPartActivity.class);
            startActivity(intent);
        });

        binding.ivNavCart.setOnClickListener(v -> {
            Log.d("CreateOrderActivity", "Cart navigation clicked - already here");
            // Already on create order screen
        });
    }

    private void createOrder() {
        Set<Integer> selectedPartIds = adapter.getSelectedParts();

        if (selectedPartIds.isEmpty()) {
            Toast.makeText(this, "Выберите хотя бы одну запчасть", Toast.LENGTH_SHORT).show();
            return;
        }

        String orderNumber = binding.etOrderNumber.getText().toString().trim();
        String buyerNumber = binding.etBuyerNumber.getText().toString().trim();
        String customerIdStr = binding.etCustomerId.getText().toString().trim();

        // Validation
        if (orderNumber.isEmpty()) {
            Toast.makeText(this, "Введите номер заказа", Toast.LENGTH_SHORT).show();
            return;
        }
        if (buyerNumber.isEmpty()) {
            Toast.makeText(this, "Введите номер покупателя", Toast.LENGTH_SHORT).show();
            return;
        }
        if (customerIdStr.isEmpty()) {
            Toast.makeText(this, "Введите ID клиента", Toast.LENGTH_SHORT).show();
            return;
        }

        int customerId;
        try {
            customerId = Integer.parseInt(customerIdStr);
        } catch (NumberFormatException e) {
            Toast.makeText(this, "Неверный ID клиента", Toast.LENGTH_SHORT).show();
            return;
        }

        // Create order object
        Order order = new Order();
        order.setCustomer_id(customerId);
        order.setPart("Заказ запчастей"); // Generic name for multi-part orders
        order.setOrder_number(orderNumber);
        order.setBuyer_number(buyerNumber);

        // Create order items from selected parts
        List<OrderItem> items = new ArrayList<>();
        for (Integer partId : selectedPartIds) {
            OrderItem item = new OrderItem();
            item.setPart_id(partId);
            item.setQuantity(1); // Default quantity, can be made configurable later
            items.add(item);
        }

        order.setItems(items);

        // Send to API
        ApiService apiService = RetrofitClient.getApiService();
        Call<Order> call = apiService.createOrder(order);

        call.enqueue(new Callback<Order>() {
            @Override
            public void onResponse(Call<Order> call, Response<Order> response) {
                if (response.isSuccessful() && response.body() != null) {
                    Log.d("CreateOrderActivity", "Order created successfully with " + selectedPartIds.size() + " parts");
                    Toast.makeText(CreateOrderActivity.this, "Заказ создан успешно", Toast.LENGTH_SHORT).show();

                    // Return to orders list
                    Intent intent = new Intent(CreateOrderActivity.this, OrdersActivity.class);
                    intent.setFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP);
                    startActivity(intent);
                    finish();
                } else {
                    Log.e("CreateOrderActivity", "Error creating order: " + response.code());
                    Toast.makeText(CreateOrderActivity.this, "Ошибка создания заказа", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<Order> call, Throwable t) {
                Log.e("CreateOrderActivity", "Network error: " + t.getMessage());
                Toast.makeText(CreateOrderActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    public void onSelectionChanged(Set<Integer> selectedParts) {
        // Update UI based on selection changes if needed
        Log.d("CreateOrderActivity", "Selection changed: " + selectedParts.size() + " parts selected");
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        binding = null;
    }
}