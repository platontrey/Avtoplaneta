package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.text.Editable;
import android.text.TextWatcher;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.Spinner;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.BulkUpdateRequest;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class HomeFragment extends Fragment implements PartsAdapter.OnPartClickListener {

    private List<InventoryItem> allParts;
    private PartsAdapter partsAdapter;

    // Элементы панели выбора
    private LinearLayout selectionPanel;
    private TextView tvSelectionCount;
    private Button btnSelectAll;
    private Button btnDeselectAll;
    private Button btnBulkOrder;
    private Button btnBulkEdit;
    private Button btnBulkDelete;
    private Button btnCancelSelection;

    // Фильтры
    private String currentSearch = "";
    private String currentCategory = "";
    private String currentBrand = "";
    private String currentModel = "";
    private String currentLocation = "";
    private String currentStatus = "";
    private String currentHasPhoto = "";
    private int currentPage = 1;
    private final int PAGE_SIZE = 20;
    private boolean isLoading = false;
    private boolean hasMorePages = true;

    // Режим множественного выбора
    private boolean isSelectionMode = false;
    private Set<Integer> selectedParts = new HashSet<>();

    @Nullable
    @Override
    public View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container, @Nullable Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.fragment_home, container, false);

        // Инициализация данных
        allParts = new ArrayList<>();

        // Настройка адаптеров
        setupAdapters(view);

        // Настройка UI
        setupUI(view);

        // Загрузка данных
        loadPartsData();

        return view;
    }

    private void setupAdapters(View view) {
        partsAdapter = new PartsAdapter(allParts, this);
        RecyclerView partsRecyclerView = view.findViewById(R.id.partsRecyclerView);
        partsRecyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        partsRecyclerView.setAdapter(partsAdapter);
    }

    private void setupUI(View view) {
        // Инициализация элементов панели выбора
        selectionPanel = view.findViewById(R.id.selectionPanel);
        tvSelectionCount = view.findViewById(R.id.tvSelectionCount);
        btnSelectAll = view.findViewById(R.id.btnSelectAll);
        btnDeselectAll = view.findViewById(R.id.btnDeselectAll);
        btnBulkOrder = view.findViewById(R.id.btnBulkOrder);
        btnBulkEdit = view.findViewById(R.id.btnBulkEdit);
        btnBulkDelete = view.findViewById(R.id.btnBulkDelete);
        btnCancelSelection = view.findViewById(R.id.btnCancelSelection);

        // Настройка обработчиков кнопок панели выбора
        btnSelectAll.setOnClickListener(v -> selectAllParts());
        btnDeselectAll.setOnClickListener(v -> deselectAllParts());
        btnBulkOrder.setOnClickListener(v -> performBulkOrder());
        btnBulkEdit.setOnClickListener(v -> performBulkEdit());
        btnBulkDelete.setOnClickListener(v -> performBulkDelete());
        btnCancelSelection.setOnClickListener(v -> exitSelectionMode());

        // Настройка поиска
        EditText searchEditText = view.findViewById(R.id.searchEditText);
        searchEditText.addTextChangedListener(new TextWatcher() {
            @Override
            public void beforeTextChanged(CharSequence s, int start, int count, int after) {}

            @Override
            public void onTextChanged(CharSequence s, int start, int before, int count) {
                filterParts(s.toString());
            }

            @Override
            public void afterTextChanged(Editable s) {}
        });

        // Настройка фильтров
        View filterIcon = view.findViewById(R.id.filterIcon);
        filterIcon.setOnClickListener(v -> showFiltersDialog());

        // Настройка бесконечной прокрутки
        RecyclerView partsRecyclerView = view.findViewById(R.id.partsRecyclerView);
        partsRecyclerView.addOnScrollListener(new RecyclerView.OnScrollListener() {
            @Override
            public void onScrolled(RecyclerView recyclerView, int dx, int dy) {
                super.onScrolled(recyclerView, dx, dy);
                LinearLayoutManager layoutManager = (LinearLayoutManager) recyclerView.getLayoutManager();
                if (layoutManager != null) {
                    int visibleItemCount = layoutManager.getChildCount();
                    int totalItemCount = layoutManager.getItemCount();
                    int firstVisibleItemPosition = layoutManager.findFirstVisibleItemPosition();

                    if (!isLoading && hasMorePages) {
                        if ((visibleItemCount + firstVisibleItemPosition) >= totalItemCount
                                && firstVisibleItemPosition >= 0) {
                            loadMoreParts();
                        }
                    }
                }
            }
        });
    }

    private void loadPartsData() {
        loadPartsData(false);
    }

    private void loadPartsData(boolean loadMore) {
        if (isLoading) return;

        isLoading = true;
        Log.d("HomeFragment", "Loading parts data, loadMore: " + loadMore + ", page: " + currentPage);

        if (!loadMore) {
            currentPage = 1;
            allParts.clear();
        }

        ApiService apiService = RetrofitClient.getApiService();
        Call<List<InventoryItem>> call = apiService.getInventory(
            currentSearch.isEmpty() ? null : currentSearch,
            currentCategory.isEmpty() ? null : currentCategory,
            currentBrand.isEmpty() ? null : currentBrand,
            currentModel.isEmpty() ? null : currentModel,
            currentLocation.isEmpty() ? null : currentLocation,
            currentStatus.isEmpty() ? null : currentStatus,
            currentHasPhoto.isEmpty() ? null : currentHasPhoto,
            PAGE_SIZE,
            currentPage
        );

        call.enqueue(new Callback<List<InventoryItem>>() {
            @Override
            public void onResponse(Call<List<InventoryItem>> call, Response<List<InventoryItem>> response) {
                isLoading = false;
                if (response.isSuccessful() && response.body() != null) {
                    List<InventoryItem> newParts = response.body();
                    if (loadMore) {
                        allParts.addAll(newParts);
                    } else {
                        allParts.clear();
                        allParts.addAll(newParts);
                    }

                    // Проверяем, есть ли еще страницы
                    hasMorePages = newParts.size() == PAGE_SIZE;
                    if (hasMorePages) {
                        currentPage++;
                    }

                    updateUI();
                    Log.d("HomeFragment", "Parts loaded: " + allParts.size() + ", hasMore: " + hasMorePages);
                } else {
                    Log.e("HomeFragment", "Error loading parts: " + response.code());
                    Toast.makeText(getContext(), "Ошибка загрузки данных: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<List<InventoryItem>> call, Throwable t) {
                isLoading = false;
                Log.e("HomeFragment", "Network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void updateUI() {
        partsAdapter.updateParts(allParts);

        // Обновляем статистику
        TextView tvTotalParts = getView().findViewById(R.id.tvTotalParts);
        TextView tvAvailableParts = getView().findViewById(R.id.tvAvailableParts);

        int totalParts = allParts.size();
        int availableParts = (int) allParts.stream().filter(InventoryItem::isStatus).count();

        tvTotalParts.setText(String.valueOf(totalParts));
        tvAvailableParts.setText(String.valueOf(availableParts));
    }

    private void filterParts(String query) {
        currentSearch = query != null ? query.trim() : "";
        loadPartsData();
    }

    private void applyFilters(String category, String brand, String model, String location, String status, String hasPhoto) {
        currentCategory = category != null ? category : "";
        currentBrand = brand != null ? brand : "";
        currentModel = model != null ? model : "";
        currentLocation = location != null ? location : "";
        currentStatus = status != null ? status : "";
        currentHasPhoto = hasPhoto != null ? hasPhoto : "";
        loadPartsData();
    }

    private void loadMoreParts() {
        if (hasMorePages && !isLoading) {
            loadPartsData(true);
        }
    }

    private void showFiltersDialog() {
        // Создаем диалог фильтров
        View dialogView = LayoutInflater.from(getContext()).inflate(R.layout.dialog_filters, null);

        // Находим элементы
        EditText etCategory = dialogView.findViewById(R.id.etCategory);
        EditText etBrand = dialogView.findViewById(R.id.etBrand);
        EditText etModel = dialogView.findViewById(R.id.etModel);
        EditText etLocation = dialogView.findViewById(R.id.etLocation);
        Spinner spStatus = dialogView.findViewById(R.id.spStatus);
        Spinner spHasPhoto = dialogView.findViewById(R.id.spHasPhoto);
        Button btnApply = dialogView.findViewById(R.id.btnApplyFilters);
        Button btnReset = dialogView.findViewById(R.id.btnResetFilters);

        // Устанавливаем текущие значения
        etCategory.setText(currentCategory);
        etBrand.setText(currentBrand);
        etModel.setText(currentModel);
        etLocation.setText(currentLocation);

        // Настраиваем спиннеры
        ArrayAdapter<CharSequence> statusAdapter = ArrayAdapter.createFromResource(getContext(),
            R.array.status_options, android.R.layout.simple_spinner_item);
        statusAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        spStatus.setAdapter(statusAdapter);
        if ("true".equals(currentStatus)) spStatus.setSelection(1);
        else if ("false".equals(currentStatus)) spStatus.setSelection(2);

        ArrayAdapter<CharSequence> photoAdapter = ArrayAdapter.createFromResource(getContext(),
            R.array.photo_options, android.R.layout.simple_spinner_item);
        photoAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        spHasPhoto.setAdapter(photoAdapter);
        if ("true".equals(currentHasPhoto)) spHasPhoto.setSelection(1);
        else if ("false".equals(currentHasPhoto)) spHasPhoto.setSelection(2);

        androidx.appcompat.app.AlertDialog dialog = new androidx.appcompat.app.AlertDialog.Builder(getContext())
            .setTitle("Фильтры")
            .setView(dialogView)
            .setNegativeButton("Отмена", null)
            .create();

        btnApply.setOnClickListener(v -> {
            String category = etCategory.getText().toString().trim();
            String brand = etBrand.getText().toString().trim();
            String model = etModel.getText().toString().trim();
            String location = etLocation.getText().toString().trim();
            String status = "";
            if (spStatus.getSelectedItemPosition() == 1) status = "true";
            else if (spStatus.getSelectedItemPosition() == 2) status = "false";
            String hasPhoto = "";
            if (spHasPhoto.getSelectedItemPosition() == 1) hasPhoto = "true";
            else if (spHasPhoto.getSelectedItemPosition() == 2) hasPhoto = "false";

            applyFilters(category, brand, model, location, status, hasPhoto);
            dialog.dismiss();
        });

        btnReset.setOnClickListener(v -> {
            applyFilters("", "", "", "", "", "");
            etCategory.setText("");
            etBrand.setText("");
            etModel.setText("");
            etLocation.setText("");
            spStatus.setSelection(0);
            spHasPhoto.setSelection(0);
        });

        dialog.show();
    }

    @Override
    public void onPartClick(InventoryItem part) {
        // Этот метод больше не используется, так как клик обрабатывается в адаптере
    }

    @Override
    public void onPartSelected(int partId, boolean isSelected) {
        if (isSelected) {
            selectedParts.add(partId);
        } else {
            selectedParts.remove(partId);
        }
        updateSelectionUI();
    }

    @Override
    public void onLongPress(int partId) {
        if (!isSelectionMode) {
            enterSelectionMode(partId);
        }
    }

    private void enterSelectionMode(int initialPartId) {
        isSelectionMode = true;
        selectedParts.clear();
        selectedParts.add(initialPartId);
        partsAdapter.setSelectionMode(true);
        updateSelectionUI();
    }

    private void exitSelectionMode() {
        isSelectionMode = false;
        selectedParts.clear();
        partsAdapter.setSelectionMode(false);
        updateSelectionUI();
    }

    private void updateSelectionUI() {
        if (isSelectionMode) {
            selectionPanel.setVisibility(View.VISIBLE);
            tvSelectionCount.setText(String.format("Выбрано: %d из %d", selectedParts.size(), allParts.size()));

            // Показываем кнопки массовых операций только если выбраны элементы
            boolean hasSelection = !selectedParts.isEmpty();
            btnBulkOrder.setVisibility(hasSelection ? View.VISIBLE : View.GONE);
            btnBulkEdit.setVisibility(hasSelection ? View.VISIBLE : View.GONE);
            btnBulkDelete.setVisibility(hasSelection ? View.VISIBLE : View.GONE);

            // Обновляем текст кнопки "Выбрать все"
            if (selectedParts.size() == allParts.size()) {
                btnSelectAll.setText("Снять все");
            } else {
                btnSelectAll.setText("Выбрать все");
            }
        } else {
            selectionPanel.setVisibility(View.GONE);
        }

        Log.d("HomeFragment", "Selection mode: " + isSelectionMode + ", selected: " + selectedParts.size());
    }

    private void selectAllParts() {
        if (selectedParts.size() == allParts.size()) {
            // Если все уже выбраны, снимаем выбор
            selectedParts.clear();
        } else {
            // Выбираем все
            selectedParts.clear();
            for (InventoryItem part : allParts) {
                selectedParts.add(part.getId());
            }
        }
        partsAdapter.notifyDataSetChanged();
        updateSelectionUI();
    }

    private void deselectAllParts() {
        selectedParts.clear();
        partsAdapter.notifyDataSetChanged();
        updateSelectionUI();
    }

    private void performBulkOrder() {
        Log.d("HomeFragment", "Bulk order for " + selectedParts.size() + " parts");

        if (selectedParts.isEmpty()) {
            Toast.makeText(getContext(), "Не выбраны запчасти для заказа", Toast.LENGTH_SHORT).show();
            return;
        }

        // Показываем диалог подтверждения
        new androidx.appcompat.app.AlertDialog.Builder(getContext())
                .setTitle("Подтверждение заказа")
                .setMessage(String.format("Создать заказ на %d запчастей?", selectedParts.size()))
                .setPositiveButton("Создать заказ", (dialog, which) -> {
                    executeBulkOrder();
                })
                .setNegativeButton("Отмена", null)
                .show();
    }

    private void executeBulkOrder() {
        // Создаем новый заказ
        com.avtoplaneta.avtoplanetaapp.models.Order newOrder = new com.avtoplaneta.avtoplanetaapp.models.Order();
        newOrder.setSeller("Менеджер"); // Можно получить из shared preferences
        newOrder.setStatus("pending");

        ApiService apiService = RetrofitClient.getApiService();
        Call<com.avtoplaneta.avtoplanetaapp.models.Order> createOrderCall = apiService.createOrder(newOrder);

        createOrderCall.enqueue(new Callback<com.avtoplaneta.avtoplanetaapp.models.Order>() {
            @Override
            public void onResponse(Call<com.avtoplaneta.avtoplanetaapp.models.Order> call, Response<com.avtoplaneta.avtoplanetaapp.models.Order> response) {
                if (response.isSuccessful() && response.body() != null) {
                    com.avtoplaneta.avtoplanetaapp.models.Order createdOrder = response.body();
                    Log.d("HomeFragment", "Order created with ID: " + createdOrder.getId());

                    // Добавляем выбранные части в заказ
                    addPartsToOrder(createdOrder.getId());
                } else {
                    Log.e("HomeFragment", "Failed to create order: " + response.code());
                    Toast.makeText(getContext(), "Ошибка создания заказа: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<com.avtoplaneta.avtoplanetaapp.models.Order> call, Throwable t) {
                Log.e("HomeFragment", "Order creation network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void addPartsToOrder(int orderId) {
        List<Integer> partIds = new ArrayList<>(selectedParts);

        // Добавляем каждую часть в заказ
        for (int partId : partIds) {
            com.avtoplaneta.avtoplanetaapp.models.AddOrderItemRequest request = new com.avtoplaneta.avtoplanetaapp.models.AddOrderItemRequest(partId, 1); // Количество 1 по умолчанию

            ApiService apiService = RetrofitClient.getApiService();
            Call<Object> addItemCall = apiService.addOrderItem(orderId, request);

            addItemCall.enqueue(new Callback<Object>() {
                @Override
                public void onResponse(Call<Object> call, Response<Object> response) {
                    if (response.isSuccessful()) {
                        Log.d("HomeFragment", "Part " + partId + " added to order " + orderId);
                    } else {
                        Log.e("HomeFragment", "Failed to add part " + partId + " to order: " + response.code());
                    }
                }

                @Override
                public void onFailure(Call<Object> call, Throwable t) {
                    Log.e("HomeFragment", "Network error adding part to order: " + t.getMessage());
                }
            });
        }

        // Выходим из режима выбора и показываем сообщение
        exitSelectionMode();
        Toast.makeText(getContext(), "Заказ создан успешно", Toast.LENGTH_SHORT).show();

        // Можно перейти к экрану заказов
        Intent intent = new Intent(getContext(), OrdersActivity.class);
        startActivity(intent);
    }

    private void performBulkEdit() {
        Log.d("HomeFragment", "Bulk edit for " + selectedParts.size() + " parts");

        if (selectedParts.isEmpty()) {
            Toast.makeText(getContext(), "Не выбраны запчасти для редактирования", Toast.LENGTH_SHORT).show();
            return;
        }

        // Создаем диалог массового редактирования
        View dialogView = LayoutInflater.from(getContext()).inflate(R.layout.dialog_bulk_edit, null);

        EditText etCategory = dialogView.findViewById(R.id.etBulkCategory);
        EditText etBrand = dialogView.findViewById(R.id.etBulkBrand);
        EditText etModel = dialogView.findViewById(R.id.etBulkModel);
        EditText etLocation = dialogView.findViewById(R.id.etBulkLocation);
        EditText etPrice = dialogView.findViewById(R.id.etBulkPrice);
        EditText etQuantity = dialogView.findViewById(R.id.etBulkQuantity);
        EditText etDescription = dialogView.findViewById(R.id.etBulkDescription);
        Button btnCancel = dialogView.findViewById(R.id.btnBulkEditCancel);
        Button btnApply = dialogView.findViewById(R.id.btnBulkEditApply);

        androidx.appcompat.app.AlertDialog dialog = new androidx.appcompat.app.AlertDialog.Builder(getContext())
                .setView(dialogView)
                .create();

        btnCancel.setOnClickListener(v -> dialog.dismiss());

        btnApply.setOnClickListener(v -> {
            // Собираем данные для обновления
            BulkUpdateRequest.BulkUpdateFields updates = new BulkUpdateRequest.BulkUpdateFields();

            String category = etCategory.getText().toString().trim();
            if (!category.isEmpty()) updates.setCategory(category);

            String brand = etBrand.getText().toString().trim();
            if (!brand.isEmpty()) updates.setBrand(brand);

            String model = etModel.getText().toString().trim();
            if (!model.isEmpty()) updates.setModel(model);

            String location = etLocation.getText().toString().trim();
            if (!location.isEmpty()) updates.setLocation(location);

            String priceStr = etPrice.getText().toString().trim();
            if (!priceStr.isEmpty()) {
                try {
                    updates.setPrice(Double.parseDouble(priceStr));
                } catch (NumberFormatException e) {
                    Toast.makeText(getContext(), "Неверный формат цены", Toast.LENGTH_SHORT).show();
                    return;
                }
            }

            String quantityStr = etQuantity.getText().toString().trim();
            if (!quantityStr.isEmpty()) {
                try {
                    updates.setQuantity(Integer.parseInt(quantityStr));
                } catch (NumberFormatException e) {
                    Toast.makeText(getContext(), "Неверный формат количества", Toast.LENGTH_SHORT).show();
                    return;
                }
            }

            String description = etDescription.getText().toString().trim();
            if (!description.isEmpty()) updates.setDescription(description);

            // Проверяем, что хотя бы одно поле заполнено
            if (updates.getCategory() == null && updates.getBrand() == null && updates.getModel() == null &&
                updates.getLocation() == null && updates.getPrice() == null && updates.getQuantity() == null &&
                updates.getDescription() == null) {
                Toast.makeText(getContext(), "Заполните хотя бы одно поле", Toast.LENGTH_SHORT).show();
                return;
            }

            dialog.dismiss();
            executeBulkUpdate(updates);
        });

        dialog.show();
    }

    private void executeBulkUpdate(BulkUpdateRequest.BulkUpdateFields updates) {
        List<Integer> idsToUpdate = new ArrayList<>(selectedParts);
        BulkUpdateRequest request = new BulkUpdateRequest(idsToUpdate, updates);

        ApiService apiService = RetrofitClient.getApiService();
        Call<Object> call = apiService.bulkUpdateParts(request);

        call.enqueue(new Callback<Object>() {
            @Override
            public void onResponse(Call<Object> call, Response<Object> response) {
                if (response.isSuccessful()) {
                    Log.d("HomeFragment", "Bulk update successful");
                    Toast.makeText(getContext(), "Запчасти успешно обновлены", Toast.LENGTH_SHORT).show();

                    // Выходим из режима выбора и обновляем данные
                    exitSelectionMode();
                    loadPartsData();
                } else {
                    Log.e("HomeFragment", "Bulk update failed: " + response.code());
                    Toast.makeText(getContext(), "Ошибка при обновлении: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<Object> call, Throwable t) {
                Log.e("HomeFragment", "Bulk update network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void performBulkDelete() {
        Log.d("HomeFragment", "Bulk delete for " + selectedParts.size() + " parts");

        if (selectedParts.isEmpty()) {
            Toast.makeText(getContext(), "Не выбраны запчасти для удаления", Toast.LENGTH_SHORT).show();
            return;
        }

        // Показываем диалог подтверждения
        new androidx.appcompat.app.AlertDialog.Builder(getContext())
                .setTitle("Подтверждение удаления")
                .setMessage(String.format("Вы действительно хотите удалить %d запчастей?", selectedParts.size()))
                .setPositiveButton("Удалить", (dialog, which) -> {
                    executeBulkDelete();
                })
                .setNegativeButton("Отмена", null)
                .show();
    }

    private void executeBulkDelete() {
        List<Integer> idsToDelete = new ArrayList<>(selectedParts);

        ApiService apiService = RetrofitClient.getApiService();
        Call<Object> call = apiService.bulkDeleteParts(idsToDelete);

        call.enqueue(new Callback<Object>() {
            @Override
            public void onResponse(Call<Object> call, Response<Object> response) {
                if (response.isSuccessful()) {
                    Log.d("HomeFragment", "Bulk delete successful");
                    Toast.makeText(getContext(), "Запчасти успешно удалены", Toast.LENGTH_SHORT).show();

                    // Выходим из режима выбора и обновляем данные
                    exitSelectionMode();
                    loadPartsData();
                } else {
                    Log.e("HomeFragment", "Bulk delete failed: " + response.code());
                    Toast.makeText(getContext(), "Ошибка при удалении: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<Object> call, Throwable t) {
                Log.e("HomeFragment", "Bulk delete network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }
}