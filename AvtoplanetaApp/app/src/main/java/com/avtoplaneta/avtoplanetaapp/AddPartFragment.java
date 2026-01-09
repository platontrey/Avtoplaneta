package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.graphics.Bitmap;
import android.net.Uri;
import android.os.Bundle;
import android.provider.MediaStore;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Toast;

import androidx.activity.result.ActivityResultLauncher;
import androidx.activity.result.contract.ActivityResultContracts;
import androidx.core.content.FileProvider;
import androidx.fragment.app.Fragment;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.databinding.FragmentAddPartBinding;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.bumptech.glide.Glide;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

import java.io.File;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.Date;

public class AddPartFragment extends Fragment {

    private FragmentAddPartBinding binding;
    private Uri selectedImageUri;
    private String currentPhotoPath;
    private boolean isEditMode = false;
    private int editingPartId = -1;

    private final ActivityResultLauncher<Uri> cameraLauncher =
        registerForActivityResult(new ActivityResultContracts.TakePicture(),
            result -> {
                Log.d("AddPartFragment", "Camera result: " + result);
                if (result) {
                    Log.d("AddPartFragment", "Photo taken successfully, compressing and loading...");
                    compressAndLoadImage();
                } else {
                    Log.d("AddPartFragment", "Photo was not taken");
                    Toast.makeText(getContext(), "Фото не было сделано", Toast.LENGTH_SHORT).show();
                }
            });

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        binding = FragmentAddPartBinding.inflate(inflater, container, false);
        return binding.getRoot();
    }

    @Override
    public void onViewCreated(View view, Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        Log.d("AddPartFragment", "onViewCreated called");

        // Check if in edit mode
        checkEditMode();

        // Update UI for edit mode
        updateUIForEditMode();

        setupListeners();

        // Update tire specs visibility based on current category
        String currentCategory = binding.etCategory.getText().toString();
        updateTireSpecsVisibility(currentCategory);
        Log.d("AddPartFragment", "onViewCreated completed");
    }

    private void setupListeners() {
        // Photo selection - directly open camera
        binding.ivPartPhoto.setOnClickListener(v -> {
            Log.d("AddPartFragment", "Photo clicked - opening camera");
            dispatchTakePictureIntent();
        });

        // Category change listener to show/hide tire specs
        binding.etCategory.addTextChangedListener(new android.text.TextWatcher() {
            @Override
            public void beforeTextChanged(CharSequence s, int start, int count, int after) {}

            @Override
            public void onTextChanged(CharSequence s, int start, int before, int count) {
                updateTireSpecsVisibility(s.toString());
            }

            @Override
            public void afterTextChanged(android.text.Editable s) {}
        });

        // Save button
        binding.btnSave.setOnClickListener(v -> {
            savePart();
        });
    }

    private void updateTireSpecsVisibility(String category) {
        boolean isTire = "Шины".equals(category) || "Tires".equals(category);
        binding.tvTireSpecs.setVisibility(isTire ? android.view.View.VISIBLE : android.view.View.GONE);
        binding.llTireRow1.setVisibility(isTire ? android.view.View.VISIBLE : android.view.View.GONE);
        binding.llTireRow2.setVisibility(isTire ? android.view.View.VISIBLE : android.view.View.GONE);
        binding.llTireRow3.setVisibility(isTire ? android.view.View.VISIBLE : android.view.View.GONE);
    }

    private void dispatchTakePictureIntent() {
        Log.d("AddPartFragment", "dispatchTakePictureIntent called");

        // Check camera permission
        if (requireActivity().checkSelfPermission(android.Manifest.permission.CAMERA) != android.content.pm.PackageManager.PERMISSION_GRANTED) {
            Log.d("AddPartFragment", "Camera permission not granted, requesting...");
            requestPermissions(new String[]{android.Manifest.permission.CAMERA}, 100);
            return;
        }

        Log.d("AddPartFragment", "Camera permission granted, proceeding...");
        Intent takePictureIntent = new Intent(MediaStore.ACTION_IMAGE_CAPTURE);
        // Ensure that there's a camera activity to handle the intent
        if (takePictureIntent.resolveActivity(requireActivity().getPackageManager()) != null) {
            Log.d("AddPartFragment", "Camera activity found, creating file...");
            // Create the File where the photo should go
            File photoFile = null;
            try {
                photoFile = createImageFile();
                Log.d("AddPartFragment", "Photo file created: " + photoFile.getAbsolutePath());
            } catch (IOException ex) {
                // Error occurred while creating the File
                Log.e("AddPartFragment", "Error creating photo file: " + ex.getMessage());
                Toast.makeText(getContext(), "Ошибка создания файла", Toast.LENGTH_SHORT).show();
                return;
            }
            // Continue only if the File was successfully created
            if (photoFile != null) {
                Uri photoURI = FileProvider.getUriForFile(requireActivity(),
                        "com.avtoplaneta.avtoplanetaapp.fileprovider",
                        photoFile);
                Log.d("AddPartFragment", "Photo URI created: " + photoURI.toString());
                takePictureIntent.putExtra(MediaStore.EXTRA_OUTPUT, photoURI);

                // Add flags to grant URI permissions
                takePictureIntent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
                takePictureIntent.addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION);

                try {
                    Log.d("AddPartFragment", "Launching camera...");
                    cameraLauncher.launch(photoURI);
                } catch (Exception e) {
                    Log.e("AddPartFragment", "Camera launch error: " + e.getMessage());
                    Toast.makeText(getContext(), "Ошибка запуска камеры", Toast.LENGTH_SHORT).show();
                }
            }
        } else {
            Log.e("AddPartFragment", "No camera activity found");
            Toast.makeText(getContext(), "Камера недоступна", Toast.LENGTH_SHORT).show();
        }
    }

    private File createImageFile() throws IOException {
        // Create an image file name
        String timeStamp = new SimpleDateFormat("yyyyMMdd_HHmmss").format(new Date());
        String imageFileName = "JPEG_" + timeStamp + "_";
        File storageDir = requireActivity().getExternalFilesDir("Pictures");
        File image = File.createTempFile(
                imageFileName,  /* prefix */
                ".jpg",         /* suffix */
                storageDir      /* directory */
        );

        // Save a file: path for use with ACTION_VIEW intents
        currentPhotoPath = image.getAbsolutePath();
        return image;
    }

    private void compressAndLoadImage() {
        Log.d("AddPartFragment", "compressAndLoadImage called");
        try {
            // Create URI from current photo path
            File photoFile = new File(currentPhotoPath);
            Log.d("AddPartFragment", "Photo file exists: " + photoFile.exists() + ", path: " + currentPhotoPath);
            selectedImageUri = Uri.fromFile(photoFile);

            // Load the full-size image
            Log.d("AddPartFragment", "Loading bitmap from URI: " + selectedImageUri);
            Bitmap bitmap = MediaStore.Images.Media.getBitmap(requireActivity().getContentResolver(), selectedImageUri);
            Log.d("AddPartFragment", "Bitmap loaded, size: " + bitmap.getWidth() + "x" + bitmap.getHeight() + ", isRecycled: " + bitmap.isRecycled());

            // Compress the image
            Bitmap compressedBitmap = compressBitmap(bitmap, 1024, 1024); // Max 1024x1024
            Log.d("AddPartFragment", "Bitmap compressed to: " + compressedBitmap.getWidth() + "x" + compressedBitmap.getHeight() + ", isRecycled: " + compressedBitmap.isRecycled());

            // Recycle original bitmap if different
            if (compressedBitmap != bitmap && !bitmap.isRecycled()) {
                bitmap.recycle();
                Log.d("AddPartFragment", "Original bitmap recycled");
            }

            // Display the compressed image
            binding.ivPartPhoto.setImageBitmap(compressedBitmap);

            Toast.makeText(getContext(), "Фото загружено", Toast.LENGTH_SHORT).show();
            Log.d("AddPartFragment", "Photo loaded successfully");

        } catch (IOException e) {
            Log.e("AddPartFragment", "Error loading image: " + e.getMessage());
            Toast.makeText(getContext(), "Ошибка загрузки изображения", Toast.LENGTH_SHORT).show();
        }
    }

    private Bitmap compressBitmap(Bitmap bitmap, int maxWidth, int maxHeight) {
        int width = bitmap.getWidth();
        int height = bitmap.getHeight();

        float scaleWidth = ((float) maxWidth) / width;
        float scaleHeight = ((float) maxHeight) / height;
        float scale = Math.min(scaleWidth, scaleHeight);

        if (scale >= 1) {
            return bitmap; // No need to scale down
        }

        int newWidth = Math.round(width * scale);
        int newHeight = Math.round(height * scale);

        return Bitmap.createScaledBitmap(bitmap, newWidth, newHeight, true);
    }

    private void savePart() {
        Log.d("AddPartFragment", "savePart called, isEditMode: " + isEditMode);

        // Get data from fields
        String name = binding.etPartName.getText().toString().trim();
        String brand = binding.etBrand.getText().toString().trim();
        String model = binding.etModel.getText().toString().trim();
        String category = binding.etCategory.getText().toString().trim();
        String location = binding.etLocation.getText().toString().trim();
        String priceStr = binding.etPrice.getText().toString().trim();
        String quantityStr = binding.etQuantity.getText().toString().trim();
        String description = binding.etDescription.getText().toString().trim();
        String vin = binding.etVin.getText().toString().trim();

        // Specifications
        String manufacturer = binding.etManufacturer.getText().toString().trim();
        String oemCode = binding.etOemCode.getText().toString().trim();
        String color = binding.etColor.getText().toString().trim();
        String condition = binding.etCondition.getText().toString().trim();
        String transmission = binding.etTransmission.getText().toString().trim();
        String drive = binding.etDrive.getText().toString().trim();
        String frontRear = binding.etFrontRear.getText().toString().trim();
        String leftRight = binding.etLeftRight.getText().toString().trim();
        String supplierCode = binding.etSupplierCode.getText().toString().trim();
        String defect = binding.etDefect.getText().toString().trim();
        String wearPercentage = binding.etWearPercentage.getText().toString().trim();

        // Tire specifications
        String season = binding.etSeason.getText().toString().trim();
        String diameter = binding.etDiameter.getText().toString().trim();
        String width = binding.etWidth.getText().toString().trim();
        String profile = binding.etProfile.getText().toString().trim();
        String tireModel = binding.etTireModel.getText().toString().trim();
        String drilling = binding.etDrilling.getText().toString().trim();

        Log.d("AddPartFragment", "Form data - name: " + name + ", brand: " + brand + ", price: " + priceStr + ", quantity: " + quantityStr);

        // Validation
        if (name.isEmpty()) {
            binding.etPartName.setError("Введите название");
            Log.d("AddPartFragment", "Validation failed: empty name");
            return;
        }

        if (brand.isEmpty()) {
            binding.etBrand.setError("Введите бренд");
            Log.d("AddPartFragment", "Validation failed: empty brand");
            return;
        }

        if (priceStr.isEmpty()) {
            binding.etPrice.setError("Введите цену");
            Log.d("AddPartFragment", "Validation failed: empty price");
            return;
        }

        if (quantityStr.isEmpty()) {
            binding.etQuantity.setError("Введите количество");
            Log.d("AddPartFragment", "Validation failed: empty quantity");
            return;
        }

        try {
            double price = Double.parseDouble(priceStr);
            int quantity = Integer.parseInt(quantityStr);
            Log.d("AddPartFragment", "Parsed values - price: " + price + ", quantity: " + quantity);

            if (isEditMode) {
                // Update existing part
                updatePart(editingPartId, name, brand, model, category, location, price, quantity, description, vin,
                          manufacturer, oemCode, color, condition, transmission, drive, frontRear, leftRight,
                          supplierCode, defect, wearPercentage, season, diameter, width, profile, tireModel, drilling);
            } else {
                // Create new part
                saveNewPart(name, brand, model, category, location, price, quantity, description, vin,
                           manufacturer, oemCode, color, condition, transmission, drive, frontRear, leftRight,
                           supplierCode, defect, wearPercentage, season, diameter, width, profile, tireModel, drilling);
            }

        } catch (NumberFormatException e) {
            Log.e("AddPartFragment", "NumberFormatException: " + e.getMessage());
            Toast.makeText(getContext(), "Неверный формат числа", Toast.LENGTH_SHORT).show();
        }
    }

    private void checkEditMode() {
        Bundle args = getArguments();
        if (args != null) {
            isEditMode = args.getBoolean("edit_mode", false);
            Log.d("AddPartFragment", "checkEditMode: isEditMode = " + isEditMode);

            if (isEditMode) {
                editingPartId = args.getInt("part_id", -1);
                Log.d("AddPartFragment", "Editing part ID: " + editingPartId);

                // Load existing data for editing
                binding.etPartName.setText(args.getString("part_name"));
                binding.etBrand.setText(args.getString("part_brand"));
                binding.etModel.setText(args.getString("part_model"));
                binding.etCategory.setText(args.getString("part_category"));
                binding.etLocation.setText(args.getString("part_location"));
                binding.etPrice.setText(String.valueOf(args.getDouble("part_price", 0)));
                binding.etQuantity.setText(String.valueOf(args.getInt("part_quantity", 0)));
                binding.etDescription.setText(args.getString("part_description"));
                binding.etVin.setText(args.getString("part_vin"));

                // Load specifications
                binding.etManufacturer.setText(args.getString("part_manufacturer"));
                binding.etOemCode.setText(args.getString("part_oem_code"));
                binding.etColor.setText(args.getString("part_color"));
                binding.etCondition.setText(args.getString("part_condition"));
                binding.etTransmission.setText(args.getString("part_transmission"));
                binding.etDrive.setText(args.getString("part_drive"));
                binding.etFrontRear.setText(args.getString("part_front_rear"));
                binding.etLeftRight.setText(args.getString("part_left_right"));
                binding.etSupplierCode.setText(args.getString("part_supplier_code"));
                binding.etDefect.setText(args.getString("part_defect"));
                binding.etWearPercentage.setText(args.getString("part_wear_percentage"));

                // Load tire specifications
                binding.etSeason.setText(args.getString("part_season"));
                binding.etDiameter.setText(args.getString("part_diameter"));
                binding.etWidth.setText(args.getString("part_width"));
                binding.etProfile.setText(args.getString("part_profile"));
                binding.etTireModel.setText(args.getString("part_tire_model"));
                binding.etDrilling.setText(args.getString("part_drilling"));

                // Load existing photo if available
                String photoUrl = args.getString("part_photo");
                if (photoUrl != null && !photoUrl.isEmpty()) {
                    String imageUrl = "http://192.168.1.63:8080" + photoUrl;
                    Log.d("AddPartFragment", "Loading existing photo from: " + imageUrl);
                    Glide.with(this)
                            .load(imageUrl)
                            .centerCrop()
                            .into(binding.ivPartPhoto);
                } else {
                    Log.d("AddPartFragment", "No existing photo to load");
                }
            }
        }
    }

    private void updateUIForEditMode() {
        if (isEditMode) {
            binding.btnSave.setText("Обновить запчасть");
            // Change title or add indicator that we're in edit mode
        }
    }

    private void saveNewPart(String name, String brand, String model, String category, String location, double price, int quantity, String description, String vin,
                            String manufacturer, String oemCode, String color, String condition, String transmission, String drive, String frontRear, String leftRight,
                            String supplierCode, String defect, String wearPercentage, String season, String diameter, String width, String profile, String tireModel, String drilling) {
        Log.d("AddPartFragment", "saveNewPart called with name: " + name + ", brand: " + brand);

        // Create new part object
        InventoryItem newPart = new InventoryItem();
        newPart.setName(name);
        newPart.setBrand(brand);
        newPart.setModel(model);
        newPart.setCategory(category);
        newPart.setLocation(location);
        newPart.setPrice(price);
        newPart.setQuantity(quantity);
        newPart.setDescription(description);
        newPart.setStatus(true);
        newPart.setVin(vin);

        // Set specifications
        newPart.setManufacturer(manufacturer);
        newPart.setOem_code(oemCode);
        newPart.setColor(color);
        newPart.setCondition(condition);
        newPart.setTransmission(transmission);
        newPart.setDrive(drive);
        newPart.setFront_rear(frontRear);
        newPart.setLeft_right(leftRight);
        newPart.setSupplier_code(supplierCode);
        newPart.setDefect(defect);
        newPart.setWear_percentage(wearPercentage);

        // Tire specifications
        newPart.setSeason(season);
        newPart.setDiameter(diameter);
        newPart.setWidth(width);
        newPart.setProfile(profile);
        newPart.setTire_model(tireModel);
        newPart.setDrilling(drilling);

        // Make API call
        ApiService apiService = RetrofitClient.getApiService();
        Call<InventoryItem> call = apiService.addPart(newPart);
        Log.d("AddPartFragment", "API call to addPart initiated");

        call.enqueue(new Callback<InventoryItem>() {
            @Override
            public void onResponse(Call<InventoryItem> call, Response<InventoryItem> response) {
                Log.d("AddPartFragment", "API response received, isSuccessful: " + response.isSuccessful() + ", code: " + response.code());
                if (response.isSuccessful() && response.body() != null) {
                    // Success - navigate back to home
                    InventoryItem createdPart = response.body();
                    Toast.makeText(getContext(), "Новая запчасть сохранена", Toast.LENGTH_SHORT).show();
                    Log.d("AddPartFragment", "Successfully created new part: " + createdPart.getName());
                    // Navigate back to home fragment
                    if (getActivity() instanceof MainActivity) {
                        ((MainActivity) getActivity()).showFragment(new HomeFragment());
                    }
                } else {
                    Toast.makeText(getContext(), "Ошибка сохранения запчасти", Toast.LENGTH_SHORT).show();
                    Log.e("AddPartFragment", "Failed to create part: " + response.code() + ", errorBody: " + (response.errorBody() != null ? response.errorBody().toString() : "null"));
                }
            }

            @Override
            public void onFailure(Call<InventoryItem> call, Throwable t) {
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
                Log.e("AddPartFragment", "Network error: " + t.getMessage());
            }
        });
    }

    private void updatePart(int partId, String name, String brand, String model, String category, String location, double price, int quantity, String description, String vin,
                           String manufacturer, String oemCode, String color, String condition, String transmission, String drive, String frontRear, String leftRight,
                           String supplierCode, String defect, String wearPercentage, String season, String diameter, String width, String profile, String tireModel, String drilling) {
        // Create updated part object
        InventoryItem updatedPart = new InventoryItem();
        updatedPart.setId(partId);
        updatedPart.setName(name);
        updatedPart.setBrand(brand);
        updatedPart.setModel(model);
        updatedPart.setCategory(category);
        updatedPart.setLocation(location);
        updatedPart.setPrice(price);
        updatedPart.setQuantity(quantity);
        updatedPart.setDescription(description);
        updatedPart.setStatus(true);
        updatedPart.setVin(vin);

        // Set specifications
        updatedPart.setManufacturer(manufacturer);
        updatedPart.setOem_code(oemCode);
        updatedPart.setColor(color);
        updatedPart.setCondition(condition);
        updatedPart.setTransmission(transmission);
        updatedPart.setDrive(drive);
        updatedPart.setFront_rear(frontRear);
        updatedPart.setLeft_right(leftRight);
        updatedPart.setSupplier_code(supplierCode);
        updatedPart.setDefect(defect);
        updatedPart.setWear_percentage(wearPercentage);

        // Tire specifications
        updatedPart.setSeason(season);
        updatedPart.setDiameter(diameter);
        updatedPart.setWidth(width);
        updatedPart.setProfile(profile);
        updatedPart.setTire_model(tireModel);
        updatedPart.setDrilling(drilling);

        // Upload photo if available
        if (selectedImageUri != null) {
            uploadPhotoAndUpdatePart(partId, updatedPart);
        } else {
            // Make API call without photo
            ApiService apiService = RetrofitClient.getApiService();
            Call<InventoryItem> call = apiService.updatePart(partId, updatedPart);

        call.enqueue(new Callback<InventoryItem>() {
            @Override
            public void onResponse(Call<InventoryItem> call, Response<InventoryItem> response) {
                if (response.isSuccessful() && response.body() != null) {
                    // Success - navigate back to home
                    InventoryItem updatedPart = response.body();
                    Toast.makeText(getContext(), "Запчасть обновлена", Toast.LENGTH_SHORT).show();
                    Log.d("AddPartFragment", "Successfully updated part: " + updatedPart.getName());
                    // Navigate back to home fragment
                    if (getActivity() instanceof MainActivity) {
                        ((MainActivity) getActivity()).showFragment(new HomeFragment());
                    }
                } else {
                    Toast.makeText(getContext(), "Ошибка обновления запчасти", Toast.LENGTH_SHORT).show();
                    Log.e("AddPartFragment", "Failed to update part: " + response.code());
                }
            }

            @Override
            public void onFailure(Call<InventoryItem> call, Throwable t) {
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
                Log.e("AddPartFragment", "Network error: " + t.getMessage());
            }
        });
    }
}

    private void uploadPhotoAndUpdatePart(int partId, InventoryItem updatedPart) {
        try {
            // Create multipart request body
            java.io.File photoFile = new java.io.File(currentPhotoPath);
            okhttp3.RequestBody requestFile = okhttp3.RequestBody.create(photoFile, okhttp3.MediaType.parse("image/*"));
            okhttp3.MultipartBody.Part body = okhttp3.MultipartBody.Part.createFormData("photo", photoFile.getName(), requestFile);

            // Upload photo first
            ApiService apiService = RetrofitClient.getApiService();
            Call<Object> photoCall = apiService.uploadPartPhoto(partId, body);

            photoCall.enqueue(new Callback<Object>() {
                @Override
                public void onResponse(Call<Object> call, Response<Object> response) {
                    if (response.isSuccessful()) {
                        Log.d("AddPartFragment", "Photo uploaded successfully, now updating part data");
                        // Photo uploaded, now update part data
                        updatePartData(partId, updatedPart);
                    } else {
                        Toast.makeText(getContext(), "Ошибка загрузки фото", Toast.LENGTH_SHORT).show();
                        Log.e("AddPartFragment", "Failed to upload photo: " + response.code());
                    }
                }

                @Override
                public void onFailure(Call<Object> call, Throwable t) {
                    Toast.makeText(getContext(), "Ошибка сети при загрузке фото", Toast.LENGTH_SHORT).show();
                    Log.e("AddPartFragment", "Network error uploading photo: " + t.getMessage());
                }
            });
        } catch (Exception e) {
            Toast.makeText(getContext(), "Ошибка обработки фото", Toast.LENGTH_SHORT).show();
            Log.e("AddPartFragment", "Error processing photo: " + e.getMessage());
        }
    }

    private void updatePartData(int partId, InventoryItem updatedPart) {
        // Make API call to update part data
        ApiService apiService = RetrofitClient.getApiService();
        Call<InventoryItem> call = apiService.updatePart(partId, updatedPart);

        call.enqueue(new Callback<InventoryItem>() {
            @Override
            public void onResponse(Call<InventoryItem> call, Response<InventoryItem> response) {
                if (response.isSuccessful() && response.body() != null) {
                    // Success - navigate back to home
                    InventoryItem updatedPart = response.body();
                    Toast.makeText(getContext(), "Запчасть обновлена с фото", Toast.LENGTH_SHORT).show();
                    Log.d("AddPartFragment", "Successfully updated part with photo: " + updatedPart.getName());
                    // Navigate back to home fragment
                    if (getActivity() instanceof MainActivity) {
                        ((MainActivity) getActivity()).showFragment(new HomeFragment());
                    }
                } else {
                    Toast.makeText(getContext(), "Ошибка обновления запчасти", Toast.LENGTH_SHORT).show();
                    Log.e("AddPartFragment", "Failed to update part: " + response.code());
                }
            }

            @Override
            public void onFailure(Call<InventoryItem> call, Throwable t) {
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
                Log.e("AddPartFragment", "Network error: " + t.getMessage());
            }
        });
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode == 100) {
            if (grantResults.length > 0 && grantResults[0] == android.content.pm.PackageManager.PERMISSION_GRANTED) {
                dispatchTakePictureIntent();
            } else {
                Toast.makeText(getContext(), "Разрешение на камеру отклонено", Toast.LENGTH_SHORT).show();
            }
        }
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();
        Log.d("AddPartFragment", "onDestroyView called, cleaning up resources");

        // Clear image view to prevent memory leaks
        if (binding != null && binding.ivPartPhoto != null) {
            binding.ivPartPhoto.setImageBitmap(null);
            Log.d("AddPartFragment", "ImageView bitmap cleared");
        }

        // Clear any cached bitmaps if needed
        if (selectedImageUri != null) {
            Log.d("AddPartFragment", "Selected image URI was set: " + selectedImageUri);
        }

        // Clear Glide cache for this image view to prevent memory leaks
        if (binding != null && binding.ivPartPhoto != null) {
            Glide.with(requireContext()).clear(binding.ivPartPhoto);
            Log.d("AddPartFragment", "Glide cache cleared for ImageView");
        }

        binding = null;
        Log.d("AddPartFragment", "Binding set to null");
    }
}