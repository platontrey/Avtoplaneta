package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.widget.ImageView;
import android.widget.TextView;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;

import com.avtoplaneta.avtoplanetaapp.databinding.ActivityPartDetailBinding;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.bumptech.glide.Glide;

public class PartDetailActivity extends AppCompatActivity {

    private ActivityPartDetailBinding binding;
    private InventoryItem part;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        binding = ActivityPartDetailBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        // Setup transparent status bar
        getWindow().setStatusBarColor(android.graphics.Color.TRANSPARENT);
        getWindow().getDecorView().setSystemUiVisibility(
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_STABLE |
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
        );

        // Get part data from intent
        part = (InventoryItem) getIntent().getSerializableExtra("part");
        if (part == null) {
            Toast.makeText(this, "Ошибка загрузки данных запчасти", Toast.LENGTH_SHORT).show();
            finish();
            return;
        }

        setupUI();
        displayPartDetails();
    }

    private void setupUI() {
        binding.btnBack.setOnClickListener(v -> finish());
        binding.btnEdit.setOnClickListener(v -> {
            Intent intent = new Intent(this, AddPartActivity.class);
            intent.putExtra("edit_mode", true);
            intent.putExtra("part_id", part.getId());
            intent.putExtra("part_name", part.getName());
            intent.putExtra("part_brand", part.getBrand());
            intent.putExtra("part_model", part.getModel());
            intent.putExtra("part_category", part.getCategory());
            intent.putExtra("part_location", part.getLocation());
            intent.putExtra("part_price", part.getPrice());
            intent.putExtra("part_quantity", part.getQuantity());
            intent.putExtra("part_description", part.getDescription());
            intent.putExtra("part_photo", part.getPhoto());
            startActivity(intent);
        });
    }

    private void displayPartDetails() {
        binding.tvPartName.setText(part.getName());
        binding.tvPartPrice.setText(String.format("Цена: %.2f руб.", part.getPrice()));
        binding.tvPartQuantity.setText(String.format("Количество: %d", part.getQuantity()));
        binding.tvPartStatus.setText(part.isStatus() ? "В наличии" : "Не в наличии");

        // Basic info
        if (part.getBrand() != null && !part.getBrand().isEmpty()) {
            binding.tvBrand.setText("Бренд: " + part.getBrand());
        } else {
            binding.tvBrand.setText("Бренд: Не указан");
        }

        if (part.getModel() != null && !part.getModel().isEmpty()) {
            binding.tvModel.setText("Модель: " + part.getModel());
        } else {
            binding.tvModel.setText("Модель: Не указана");
        }

        if (part.getCategory() != null && !part.getCategory().isEmpty()) {
            binding.tvCategory.setText("Категория: " + part.getCategory());
        } else {
            binding.tvCategory.setText("Категория: Не указана");
        }

        if (part.getLocation() != null && !part.getLocation().isEmpty()) {
            binding.tvLocation.setText("Местоположение: " + part.getLocation());
        } else {
            binding.tvLocation.setText("Местоположение: Не указано");
        }

        // Description
        if (part.getDescription() != null && !part.getDescription().isEmpty()) {
            binding.tvDescription.setText(part.getDescription());
        } else {
            binding.tvDescription.setText("Описание отсутствует");
        }

        // Additional characteristics
        StringBuilder characteristics = new StringBuilder();

        if (part.getVin() != null && !part.getVin().isEmpty()) {
            characteristics.append("VIN: ").append(part.getVin()).append("\n");
        }

        if (part.getBody_brand() != null && !part.getBody_brand().isEmpty()) {
            characteristics.append("Марка кузова: ").append(part.getBody_brand()).append("\n");
        }

        if (part.getEngine_brand() != null && !part.getEngine_brand().isEmpty()) {
            characteristics.append("Марка двигателя: ").append(part.getEngine_brand()).append("\n");
        }

        if (part.getCar_release_date() != null && !part.getCar_release_date().isEmpty()) {
            characteristics.append("Год выпуска: ").append(part.getCar_release_date()).append("\n");
        }

        if (part.getTransmission() != null && !part.getTransmission().isEmpty()) {
            characteristics.append("Трансмиссия: ").append(part.getTransmission()).append("\n");
        }

        if (part.getDrive() != null && !part.getDrive().isEmpty()) {
            characteristics.append("Привод: ").append(part.getDrive()).append("\n");
        }

        if (characteristics.length() > 0) {
            binding.tvCharacteristics.setText(characteristics.toString().trim());
        } else {
            binding.tvCharacteristics.setText("Дополнительные характеристики отсутствуют");
        }

        // Load photo if available
        if (part.getPhoto() != null && !part.getPhoto().isEmpty()) {
            String imageUrl = "http://192.168.1.63:8080" + part.getPhoto();
            Glide.with(this)
                    .load(imageUrl)
                    .centerCrop()
                    .into(binding.ivPartPhoto);
        } else {
            binding.ivPartPhoto.setImageResource(R.drawable.ic_placeholder_part);
        }
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (binding != null) {
            Glide.with(getApplicationContext()).clear(binding.ivPartPhoto);
            binding = null;
        }
    }
}