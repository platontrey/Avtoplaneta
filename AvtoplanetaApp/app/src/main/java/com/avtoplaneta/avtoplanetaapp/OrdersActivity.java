package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.ImageView;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.Order;

import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class OrdersActivity extends AppCompatActivity {

    private OrdersAdapter ordersAdapter;
    private List<Order> ordersList;
    private RecyclerView ordersRecyclerView;
    private ImageView ivBack;
    private ImageView ivAddOrder;
    private ImageView ivNavHome;
    private ImageView ivNavAdd;
    private ImageView ivNavOrders;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("OrdersActivity", "=== ORDERS ACTIVITY STARTED ===");

        setContentView(R.layout.activity_orders);

        // Initialize views
        ordersRecyclerView = findViewById(R.id.ordersRecyclerView);
        ivBack = findViewById(R.id.ivBack);
        ivAddOrder = findViewById(R.id.ivAddOrder);
        ivNavHome = findViewById(R.id.ivNavHome);
        ivNavAdd = findViewById(R.id.ivNavAdd);
        ivNavOrders = findViewById(R.id.ivNavOrders);

        // Setup RecyclerView
        setupRecyclerView();

        // Setup UI
        setupUI();

        // Load orders
        loadOrders();
    }

    private void setupRecyclerView() {
        ordersAdapter = new OrdersAdapter(this);
        ordersRecyclerView.setLayoutManager(new LinearLayoutManager(this));
        ordersRecyclerView.setAdapter(ordersAdapter);
    }

    private void setupUI() {
        // Back button
        ivBack.setOnClickListener(v -> {
            Log.d("OrdersActivity", "Back button clicked");
            finish();
        });

        // Add order button
        ivAddOrder.setOnClickListener(v -> {
            Log.d("OrdersActivity", "Add order button clicked");
            Intent intent = new Intent(OrdersActivity.this, CreateOrderActivity.class);
            startActivity(intent);
        });

        // Bottom navigation
        ivNavHome.setOnClickListener(v -> {
            Log.d("OrdersActivity", "Home navigation clicked");
            Intent intent = new Intent(OrdersActivity.this, MainActivity.class);
            intent.setFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP);
            startActivity(intent);
            finish();
        });

        ivNavAdd.setOnClickListener(v -> {
            Log.d("OrdersActivity", "Add navigation clicked");
            Intent intent = new Intent(OrdersActivity.this, CreateOrderActivity.class);
            startActivity(intent);
        });

        ivNavOrders.setOnClickListener(v -> {
            Log.d("OrdersActivity", "Orders navigation clicked - already here");
            // Already on orders screen
        });
    }

    private void loadOrders() {
        Log.d("OrdersActivity", "Loading orders");
        ApiService apiService = RetrofitClient.getApiService();
        Call<List<Order>> call = apiService.getOrders();

        call.enqueue(new Callback<List<Order>>() {
            @Override
            public void onResponse(Call<List<Order>> call, Response<List<Order>> response) {
                if (response.isSuccessful() && response.body() != null) {
                    ordersList = response.body();
                    ordersAdapter.updateOrders(ordersList);
                    Log.d("OrdersActivity", "Orders loaded: " + ordersList.size());
                } else {
                    Log.e("OrdersActivity", "Error loading orders: " + response.code());
                    Toast.makeText(OrdersActivity.this, "Ошибка загрузки заказов", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<List<Order>> call, Throwable t) {
                Log.e("OrdersActivity", "Network error: " + t.getMessage());
                Toast.makeText(OrdersActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    protected void onResume() {
        super.onResume();
        // Refresh orders when returning to this activity
        loadOrders();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
    }
}