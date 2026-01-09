package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.CheckBox;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;

import java.util.HashSet;
import java.util.List;
import java.util.Set;
import timber.log.Timber;

public class PartsAdapter extends RecyclerView.Adapter<PartsAdapter.PartViewHolder> {

    private List<InventoryItem> partsList;
    private OnPartClickListener listener;
    private boolean isSelectionMode = false;
    private Set<Integer> selectedParts = new HashSet<>();

    public interface OnPartClickListener {
        void onPartClick(InventoryItem part);
        void onPartSelected(int partId, boolean isSelected);
        void onLongPress(int partId);
    }

    public PartsAdapter(List<InventoryItem> partsList, OnPartClickListener listener) {
        this.partsList = partsList;
        this.listener = listener;
    }

    public void updateParts(List<InventoryItem> newParts) {
        this.partsList = newParts;
        notifyDataSetChanged();
    }

    public void setSelectionMode(boolean selectionMode) {
        this.isSelectionMode = selectionMode;
        if (!selectionMode) {
            selectedParts.clear();
        }
        notifyDataSetChanged();
    }

    public boolean isSelectionMode() {
        return isSelectionMode;
    }

    public Set<Integer> getSelectedParts() {
        return new HashSet<>(selectedParts);
    }

    public void selectAll() {
        selectedParts.clear();
        for (InventoryItem part : partsList) {
            selectedParts.add(part.getId());
        }
        notifyDataSetChanged();
    }

    public void deselectAll() {
        selectedParts.clear();
        notifyDataSetChanged();
    }

    public void toggleSelection(int partId) {
        if (selectedParts.contains(partId)) {
            selectedParts.remove(partId);
        } else {
            selectedParts.add(partId);
        }
        notifyDataSetChanged();
    }

    @NonNull
    @Override
    public PartViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_part, parent, false);
        return new PartViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull PartViewHolder holder, int position) {
        InventoryItem part = partsList.get(position);
        holder.bind(part, listener, isSelectionMode, selectedParts.contains(part.getId()));
    }

    @Override
    public int getItemCount() {
        return partsList.size();
    }

    static class PartViewHolder extends RecyclerView.ViewHolder {
        private CheckBox cbPartSelected;
        private TextView tvPartName;
        private TextView tvPartDescription;
        private TextView tvPartPrice;
        private TextView tvPartQuantity;

        public PartViewHolder(@NonNull View itemView) {
            super(itemView);
            cbPartSelected = itemView.findViewById(R.id.cbPartSelected);
            tvPartName = itemView.findViewById(R.id.tvPartName);
            tvPartDescription = itemView.findViewById(R.id.tvPartDescription);
            tvPartPrice = itemView.findViewById(R.id.tvPartPrice);
            tvPartQuantity = itemView.findViewById(R.id.tvPartQuantity);
        }

        public void bind(InventoryItem part, OnPartClickListener listener, boolean isSelectionMode, boolean isSelected) {
            tvPartName.setText(part.getName());

            // Формируем описание с дополнительной информацией
            StringBuilder description = new StringBuilder();
            if (part.getDescription() != null && !part.getDescription().isEmpty()) {
                description.append(part.getDescription());
            } else {
                description.append("Нет описания");
            }

            // Добавляем бренд и модель
            if (part.getBrand() != null && !part.getBrand().isEmpty()) {
                description.append("\nБренд: ").append(part.getBrand());
            }
            if (part.getModel() != null && !part.getModel().isEmpty()) {
                description.append(", Модель: ").append(part.getModel());
            }
            if (part.getCategory() != null && !part.getCategory().isEmpty()) {
                description.append("\nКатегория: ").append(part.getCategory());
            }
            if (part.getLocation() != null && !part.getLocation().isEmpty()) {
                description.append(", Местоположение: ").append(part.getLocation());
            }

            tvPartDescription.setText(description.toString());
            tvPartPrice.setText(String.format("Цена: %.2f руб.", part.getPrice()));
            tvPartQuantity.setText(String.format("Количество: %d", part.getQuantity()));

            // Настройка чекбокса для режима выбора
            if (cbPartSelected != null) {
                if (isSelectionMode) {
                    cbPartSelected.setVisibility(View.VISIBLE);
                    cbPartSelected.setChecked(isSelected);
                    cbPartSelected.setOnCheckedChangeListener((buttonView, isChecked) -> {
                        if (listener != null) {
                            listener.onPartSelected(part.getId(), isChecked);
                        }
                    });
                } else {
                    cbPartSelected.setVisibility(View.GONE);
                }
            }

            itemView.setOnClickListener(v -> {
                if (isSelectionMode) {
                    // В режиме выбора клик переключает чекбокс
                    if (cbPartSelected != null) {
                        cbPartSelected.setChecked(!cbPartSelected.isChecked());
                    }
                } else {
                    // Открываем детальный вид
                    Timber.d("Starting PartDetailActivity for part: %s (id: %d)", part.getName(), part.getId());
                    Intent intent = new Intent(itemView.getContext(), PartDetailActivity.class);
                    intent.putExtra("part", part);
                    itemView.getContext().startActivity(intent);
                }
            });

            itemView.setOnLongClickListener(v -> {
                if (listener != null) {
                    if (isSelectionMode) {
                        // В режиме выбора долгое нажатие не делает ничего особого
                        return true;
                    } else {
                        // В обычном режиме долгое нажатие активирует режим выбора
                        listener.onLongPress(part.getId());
                    }
                }
                return true;
            });
        }

        // Устаревший метод для обратной совместимости
        public void bind(InventoryItem part, OnPartClickListener listener) {
            bind(part, listener, false, false);
        }
    }
}