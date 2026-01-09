package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.ActivityAddCarBinding;
import com.avtoplaneta.avtoplanetaapp.models.Car;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class AddCarActivity extends AppCompatActivity {

    private ActivityAddCarBinding binding;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.d("AddCarActivity", "=== ADD CAR ACTIVITY STARTED ===");

        binding = ActivityAddCarBinding.inflate(getLayoutInflater());
        setContentView(binding.getRoot());

        setupUI();
    }

    private void setupUI() {
        // Back button
        binding.ivBack.setOnClickListener(v -> {
            Log.d("AddCarActivity", "Back button clicked");
            finish();
        });

        // Save button
        binding.btnSave.setOnClickListener(v -> {
            Log.d("AddCarActivity", "Save button clicked");
            saveCar();
        });
    }

    private void saveCar() {
        // Get data from fields
        String brand = binding.etBrand.getText().toString().trim();
        String model = binding.etModel.getText().toString().trim();
        String yearStr = binding.etYear.getText().toString().trim();
        String vin = binding.etVin.getText().toString().trim();
        String mileageStr = binding.etMileage.getText().toString().trim();
        String color = binding.etColor.getText().toString().trim();
        String engineType = binding.etEngineType.getText().toString().trim();
        String transmission = binding.etTransmission.getText().toString().trim();
        String driveType = binding.etDriveType.getText().toString().trim();
        String bodyType = binding.etBodyType.getText().toString().trim();
        String registrationNumber = binding.etRegistrationNumber.getText().toString().trim();

        // Validation
        if (brand.isEmpty()) {
            binding.etBrand.setError("Введите марку");
            return;
        }

        if (model.isEmpty()) {
            binding.etModel.setError("Введите модель");
            return;
        }

        if (yearStr.isEmpty()) {
            binding.etYear.setError("Введите год");
            return;
        }

        int year;
        try {
            year = Integer.parseInt(yearStr);
        } catch (NumberFormatException e) {
            binding.etYear.setError("Неверный год");
            return;
        }

        int mileage = 0;
        if (!mileageStr.isEmpty()) {
            try {
                mileage = Integer.parseInt(mileageStr);
            } catch (NumberFormatException e) {
                binding.etMileage.setError("Неверный пробег");
                return;
            }
        }

        // Create car object
        Car car = new Car();
        car.setBrand(brand);
        car.setModel(model);
        car.setYear(year);
        car.setVin(vin);
        car.setMileage(mileage);
        car.setColor(color);
        car.setEngine_type(engineType);
        car.setTransmission(transmission);
        car.setDrive_type(driveType);
        car.setBody_type(bodyType);
        car.setRegistration_number(registrationNumber);

        // Send to API
        ApiService apiService = RetrofitClient.getApiService();
        Call<Car> call = apiService.addCar(car);

        call.enqueue(new Callback<Car>() {
            @Override
            public void onResponse(Call<Car> call, Response<Car> response) {
                if (response.isSuccessful() && response.body() != null) {
                    Log.d("AddCarActivity", "Car added successfully");
                    Toast.makeText(AddCarActivity.this, "Автомобиль добавлен успешно", Toast.LENGTH_SHORT).show();

                    // Return to previous screen
                    finish();
                } else {
                    Log.e("AddCarActivity", "Error adding car: " + response.code());
                    Toast.makeText(AddCarActivity.this, "Ошибка добавления автомобиля", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<Car> call, Throwable t) {
                Log.e("AddCarActivity", "Network error: " + t.getMessage());
                Toast.makeText(AddCarActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        binding = null;
    }
}