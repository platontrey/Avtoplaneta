package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.Order;

import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class OrdersFragment extends Fragment {

    private OrdersAdapter ordersAdapter;
    private List<Order> ordersList;
    private RecyclerView ordersRecyclerView;
    private ImageView ivAddOrder;

    @Nullable
    @Override
    public View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container, @Nullable Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.fragment_orders, container, false);

        // Initialize views
        ordersRecyclerView = view.findViewById(R.id.ordersRecyclerView);
        ivAddOrder = view.findViewById(R.id.ivAddOrder);

        // Setup RecyclerView
        setupRecyclerView();

        // Setup UI
        setupUI();

        // Load orders
        loadOrders();

        return view;
    }

    private void setupRecyclerView() {
        ordersAdapter = new OrdersAdapter(getContext());
        ordersRecyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        ordersRecyclerView.setAdapter(ordersAdapter);
    }

    private void setupUI() {
        // Add order button
        ivAddOrder.setOnClickListener(v -> {
            Log.d("OrdersFragment", "Add order button clicked");
            Intent intent = new Intent(getContext(), CreateOrderActivity.class);
            startActivity(intent);
        });
    }

    private void loadOrders() {
        Log.d("OrdersFragment", "Loading orders");
        ApiService apiService = RetrofitClient.getApiService();
        Call<List<Order>> call = apiService.getOrders();

        call.enqueue(new Callback<List<Order>>() {
            @Override
            public void onResponse(Call<List<Order>> call, Response<List<Order>> response) {
                if (response.isSuccessful() && response.body() != null) {
                    ordersList = response.body();
                    ordersAdapter.updateOrders(ordersList);
                    Log.d("OrdersFragment", "Orders loaded: " + ordersList.size());
                } else {
                    Log.e("OrdersFragment", "Error loading orders: " + response.code());
                    Toast.makeText(getContext(), "Ошибка загрузки заказов", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<List<Order>> call, Throwable t) {
                Log.e("OrdersFragment", "Network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    public void onResume() {
        super.onResume();
        // Refresh orders when returning to this fragment
        loadOrders();
    }
}