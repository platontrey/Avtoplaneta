package com.avtoplaneta.avtoplanetaapp;

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

public class OrderPartsAdapter extends RecyclerView.Adapter<OrderPartsAdapter.PartViewHolder> {

    private List<InventoryItem> partsList;
    private Set<Integer> selectedParts = new HashSet<>();
    private OnSelectionChangedListener listener;

    public interface OnSelectionChangedListener {
        void onSelectionChanged(Set<Integer> selectedParts);
    }

    public OrderPartsAdapter(List<InventoryItem> partsList, OnSelectionChangedListener listener) {
        this.partsList = partsList;
        this.listener = listener;
    }

    @NonNull
    @Override
    public PartViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_order_part, parent, false);
        return new PartViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull PartViewHolder holder, int position) {
        InventoryItem part = partsList.get(position);
        holder.bind(part, position);
    }

    @Override
    public int getItemCount() {
        return partsList.size();
    }

    public Set<Integer> getSelectedParts() {
        return selectedParts;
    }

    public List<InventoryItem> getSelectedPartItems() {
        return partsList.stream()
                .filter(part -> selectedParts.contains(part.getId()))
                .collect(java.util.stream.Collectors.toList());
    }

    class PartViewHolder extends RecyclerView.ViewHolder {
        private CheckBox checkBox;
        private TextView tvPartName;
        private TextView tvPartDescription;
        private TextView tvPartPrice;
        private TextView tvPartQuantity;

        public PartViewHolder(@NonNull View itemView) {
            super(itemView);
            checkBox = itemView.findViewById(R.id.checkBox);
            tvPartName = itemView.findViewById(R.id.tvPartName);
            tvPartDescription = itemView.findViewById(R.id.tvPartDescription);
            tvPartPrice = itemView.findViewById(R.id.tvPartPrice);
            tvPartQuantity = itemView.findViewById(R.id.tvPartQuantity);
        }

        public void bind(InventoryItem part, int position) {
            tvPartName.setText(part.getName());
            tvPartDescription.setText(part.getDescription() != null ? part.getDescription() : "Нет описания");
            tvPartPrice.setText(String.format("Цена: %.2f руб.", part.getPrice()));
            tvPartQuantity.setText(String.format("Количество: %d", part.getQuantity()));

            checkBox.setChecked(selectedParts.contains(part.getId()));

            checkBox.setOnCheckedChangeListener((buttonView, isChecked) -> {
                if (isChecked) {
                    selectedParts.add(part.getId());
                } else {
                    selectedParts.remove(part.getId());
                }
                if (listener != null) {
                    listener.onSelectionChanged(selectedParts);
                }
            });

            itemView.setOnClickListener(v -> {
                checkBox.setChecked(!checkBox.isChecked());
            });
        }
    }
}