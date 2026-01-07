package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.LinearLayout;
import android.widget.TextView;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.cardview.widget.CardView;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.CategoryCount;
import com.avtoplaneta.avtoplanetaapp.models.MonthlySales;
import com.avtoplaneta.avtoplanetaapp.models.StatisticsResponse;

import java.text.NumberFormat;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class StatisticsActivity extends AppCompatActivity {

    private static final String PREFS_NAME = "AvtoplanetaPrefs";
    private static final String KEY_IS_LOGGED_IN = "is_logged_in";

    // UI elements
    private TextView tvTotalParts, tvTotalQuantity, tvTotalValue, tvTotalEarnings;
    private RecyclerView rvCategories, rvMonthlySales;
    private LinearLayout loadingLayout, contentLayout;

    // Data
    private List<CategoryCount> categories = new ArrayList<>();
    private List<MonthlySales> monthlySales = new ArrayList<>();

    // Adapters
    private CategoryAdapter categoryAdapter;
    private MonthlySalesAdapter monthlySalesAdapter;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("StatisticsActivity", "=== STATISTICS ACTIVITY STARTED ===");

        // Check authentication
        SharedPreferences sharedPreferences = getSharedPreferences(PREFS_NAME, MODE_PRIVATE);
        boolean isLoggedIn = sharedPreferences.getBoolean(KEY_IS_LOGGED_IN, false);

        if (!isLoggedIn) {
            Intent intent = new Intent(StatisticsActivity.this, LoginActivity.class);
            intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TASK);
            startActivity(intent);
            finish();
            return;
        }

        setContentView(R.layout.activity_statistics);

        initializeViews();
        setupAdapters();
        loadStatistics();
    }

    private void initializeViews() {
        // Summary cards
        tvTotalParts = findViewById(R.id.tvTotalParts);
        tvTotalQuantity = findViewById(R.id.tvTotalQuantity);
        tvTotalValue = findViewById(R.id.tvTotalValue);
        tvTotalEarnings = findViewById(R.id.tvTotalEarnings);

        // RecyclerViews
        rvCategories = findViewById(R.id.rvCategories);
        rvMonthlySales = findViewById(R.id.rvMonthlySales);

        // Loading and content layouts
        loadingLayout = findViewById(R.id.loadingLayout);
        contentLayout = findViewById(R.id.contentLayout);

        // Back button
        findViewById(R.id.btnBack).setOnClickListener(v -> finish());
    }

    private void setupAdapters() {
        // Categories adapter
        categoryAdapter = new CategoryAdapter(categories);
        rvCategories.setLayoutManager(new LinearLayoutManager(this));
        rvCategories.setAdapter(categoryAdapter);

        // Monthly sales adapter
        monthlySalesAdapter = new MonthlySalesAdapter(monthlySales);
        rvMonthlySales.setLayoutManager(new LinearLayoutManager(this));
        rvMonthlySales.setAdapter(monthlySalesAdapter);
    }

    private void loadStatistics() {
        showLoading(true);

        ApiService apiService = RetrofitClient.getApiService();
        Call<StatisticsResponse> call = apiService.getStatistics();

        call.enqueue(new Callback<StatisticsResponse>() {
            @Override
            public void onResponse(Call<StatisticsResponse> call, Response<StatisticsResponse> response) {
                showLoading(false);

                if (response.isSuccessful() && response.body() != null) {
                    StatisticsResponse stats = response.body();
                    updateUI(stats);
                    Log.d("StatisticsActivity", "Statistics loaded successfully");
                } else {
                    Log.e("StatisticsActivity", "Error loading statistics: " + response.code());
                    Toast.makeText(StatisticsActivity.this, "Ошибка загрузки статистики: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<StatisticsResponse> call, Throwable t) {
                showLoading(false);
                Log.e("StatisticsActivity", "Network error: " + t.getMessage());
                Toast.makeText(StatisticsActivity.this, "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void updateUI(StatisticsResponse stats) {
        // Update summary cards
        tvTotalParts.setText(String.valueOf(stats.getTotal_parts()));
        tvTotalQuantity.setText(String.valueOf(stats.getTotal_quantity()));
        tvTotalValue.setText(formatCurrency(stats.getTotal_value()));
        tvTotalEarnings.setText(formatCurrency(stats.getTotal_earnings()));

        // Update categories
        if (stats.getCategories() != null) {
            categories.clear();
            categories.addAll(stats.getCategories());
            categoryAdapter.notifyDataSetChanged();
        }

        // Update monthly sales
        if (stats.getMonthly_sales() != null) {
            monthlySales.clear();
            monthlySales.addAll(stats.getMonthly_sales());
            monthlySalesAdapter.notifyDataSetChanged();
        }
    }

    private void showLoading(boolean show) {
        loadingLayout.setVisibility(show ? View.VISIBLE : View.GONE);
        contentLayout.setVisibility(show ? View.GONE : View.VISIBLE);
    }

    private String formatCurrency(double value) {
        NumberFormat format = NumberFormat.getCurrencyInstance(new Locale("ru", "RU"));
        return format.format(value);
    }

    // Category Adapter
    private static class CategoryAdapter extends RecyclerView.Adapter<CategoryAdapter.ViewHolder> {

        private List<CategoryCount> categories;

        public CategoryAdapter(List<CategoryCount> categories) {
            this.categories = categories;
        }

        @Override
        public ViewHolder onCreateViewHolder(android.view.ViewGroup parent, int viewType) {
            View view = android.view.LayoutInflater.from(parent.getContext())
                    .inflate(R.layout.item_category_stat, parent, false);
            return new ViewHolder(view);
        }

        @Override
        public void onBindViewHolder(ViewHolder holder, int position) {
            CategoryCount category = categories.get(position);
            holder.tvCategoryName.setText(category.getName());
            holder.tvCategoryCount.setText(String.valueOf(category.getCount()));
        }

        @Override
        public int getItemCount() {
            return categories.size();
        }

        static class ViewHolder extends RecyclerView.ViewHolder {
            TextView tvCategoryName, tvCategoryCount;

            ViewHolder(View itemView) {
                super(itemView);
                tvCategoryName = itemView.findViewById(R.id.tvCategoryName);
                tvCategoryCount = itemView.findViewById(R.id.tvCategoryCount);
            }
        }
    }

    // Monthly Sales Adapter
    private static class MonthlySalesAdapter extends RecyclerView.Adapter<MonthlySalesAdapter.ViewHolder> {

        private List<MonthlySales> monthlySales;

        public MonthlySalesAdapter(List<MonthlySales> monthlySales) {
            this.monthlySales = monthlySales;
        }

        @Override
        public ViewHolder onCreateViewHolder(android.view.ViewGroup parent, int viewType) {
            View view = android.view.LayoutInflater.from(parent.getContext())
                    .inflate(R.layout.item_monthly_sales_stat, parent, false);
            return new ViewHolder(view);
        }

        @Override
        public void onBindViewHolder(ViewHolder holder, int position) {
            MonthlySales sales = monthlySales.get(position);
            holder.tvMonth.setText(sales.getMonth());
            holder.tvSales.setText(formatCurrency(sales.getSales()));
        }

        @Override
        public int getItemCount() {
            return monthlySales.size();
        }

        private String formatCurrency(double value) {
            NumberFormat format = NumberFormat.getCurrencyInstance(new Locale("ru", "RU"));
            return format.format(value);
        }

        static class ViewHolder extends RecyclerView.ViewHolder {
            TextView tvMonth, tvSales;

            ViewHolder(View itemView) {
                super(itemView);
                tvMonth = itemView.findViewById(R.id.tvMonth);
                tvSales = itemView.findViewById(R.id.tvSales);
            }
        }
    }
}