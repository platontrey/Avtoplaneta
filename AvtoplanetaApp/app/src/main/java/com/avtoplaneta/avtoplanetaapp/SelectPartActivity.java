package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.ActivitySelectPartBinding;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;

import java.util.ArrayList;
import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class SelectPartActivity extends AppCompatActivity implements PartsAdapter.OnPartClickListener {

    private ActivitySelectPartBinding binding;
    private PartsAdapter adapter;
    private List<InventoryItem> partsList = new ArrayList<>();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("SelectPartActivity", "=== SELECT PART ACTIVITY STARTED ===");

        binding = ActivitySelectPartBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        setupUI();
        loadParts();
    }

    private void setupUI() {
        // Back button
        binding.ivBack.setOnClickListener(v -> {
            Log.d("SelectPartActivity", "Back button clicked");
            finish();
        });

        // Setup RecyclerView
        binding.rvParts.setLayoutManager(new LinearLayoutManager(this));
        adapter = new PartsAdapter(partsList, this);
        binding.rvParts.setAdapter(adapter);
    }

    private void loadParts() {
        Log.d("SelectPartActivity", "Loading parts from API");

        ApiService apiService = RetrofitClient.getApiService();
        Call<List<InventoryItem>> call = apiService.getInventory();

        call.enqueue(new Callback<List<InventoryItem>>() {
            @Override
            public void onResponse(Call<List<InventoryItem>> call, Response<List<InventoryItem>> response) {
                if (response.isSuccessful() && response.body() != null) {
                    partsList.clear();
                    partsList.addAll(response.body());
                    adapter.notifyDataSetChanged();
                    Log.d("SelectPartActivity", "Loaded " + partsList.size() + " parts");
                } else {
                    Log.e("SelectPartActivity", "Error loading parts: " + response.code());
                    Toast.makeText(SelectPartActivity.this, "Ошибка загрузки запчастей", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<List<InventoryItem>> call, Throwable t) {
                Log.e("SelectPartActivity", "Network error: " + t.getMessage());
                Toast.makeText(SelectPartActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    public void onPartClick(InventoryItem part) {
        Log.d("SelectPartActivity", "Part selected: " + part.getName() + " (ID: " + part.getId() + ")");

        // Return selected part to CreateOrderActivity
        Intent resultIntent = new Intent();
        resultIntent.putExtra("selected_part", part);
        setResult(RESULT_OK, resultIntent);
        finish();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        binding = null;
    }
}