package com.avtoplaneta.avtoplanetaapp;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;

import java.util.List;

public class PartsAdapter extends RecyclerView.Adapter<PartsAdapter.PartViewHolder> {

    private List<InventoryItem> partsList;
    private OnPartClickListener listener;

    public interface OnPartClickListener {
        void onPartClick(InventoryItem part);
    }

    public PartsAdapter(List<InventoryItem> partsList, OnPartClickListener listener) {
        this.partsList = partsList;
        this.listener = listener;
    }

    public void updateParts(List<InventoryItem> newParts) {
        this.partsList = newParts;
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
        holder.bind(part, listener);
    }

    @Override
    public int getItemCount() {
        return partsList.size();
    }

    static class PartViewHolder extends RecyclerView.ViewHolder {
        private TextView tvPartName;
        private TextView tvPartDescription;
        private TextView tvPartPrice;
        private TextView tvPartQuantity;

        public PartViewHolder(@NonNull View itemView) {
            super(itemView);
            tvPartName = itemView.findViewById(R.id.tvPartName);
            tvPartDescription = itemView.findViewById(R.id.tvPartDescription);
            tvPartPrice = itemView.findViewById(R.id.tvPartPrice);
            tvPartQuantity = itemView.findViewById(R.id.tvPartQuantity);
        }

        public void bind(InventoryItem part, OnPartClickListener listener) {
            tvPartName.setText(part.getName());
            tvPartDescription.setText(part.getDescription() != null ? part.getDescription() : "Нет описания");
            tvPartPrice.setText(String.format("Цена: %.2f руб.", part.getPrice()));
            tvPartQuantity.setText(String.format("Количество: %d", part.getQuantity()));

            itemView.setOnClickListener(v -> {
                if (listener != null) {
                    listener.onPartClick(part);
                }
            });
        }
    }
}