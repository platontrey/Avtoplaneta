package com.avtoplaneta.avtoplanetaapp;

import android.content.Context;
import android.graphics.Color;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.models.Order;

import java.util.ArrayList;
import java.util.List;

public class OrdersAdapter extends RecyclerView.Adapter<OrdersAdapter.OrderViewHolder> {

    private List<Order> orders;
    private Context context;

    public OrdersAdapter(Context context) {
        this.context = context;
        this.orders = new ArrayList<>();
    }

    @NonNull
    @Override
    public OrderViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(context).inflate(R.layout.item_order, parent, false);
        return new OrderViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull OrderViewHolder holder, int position) {
        Order order = orders.get(position);
        holder.bind(order);
    }

    @Override
    public int getItemCount() {
        return orders.size();
    }

    public void updateOrders(List<Order> newOrders) {
        this.orders = newOrders != null ? newOrders : new ArrayList<>();
        notifyDataSetChanged();
    }

    static class OrderViewHolder extends RecyclerView.ViewHolder {
        private TextView tvOrderNumber;
        private TextView tvPart;
        private TextView tvBuyerNumber;
        private TextView tvStatus;
        private TextView tvSeller;
        private TextView tvTimeAgo;

        public OrderViewHolder(@NonNull View itemView) {
            super(itemView);
            tvOrderNumber = itemView.findViewById(R.id.tvOrderNumber);
            tvPart = itemView.findViewById(R.id.tvPart);
            tvBuyerNumber = itemView.findViewById(R.id.tvBuyerNumber);
            tvStatus = itemView.findViewById(R.id.tvStatus);
            tvSeller = itemView.findViewById(R.id.tvSeller);
            tvTimeAgo = itemView.findViewById(R.id.tvTimeAgo);
        }

        public void bind(Order order) {
            tvOrderNumber.setText("Заказ #" + order.getOrder_number());
            tvPart.setText(order.getPart());
            tvBuyerNumber.setText("Покупатель: " + order.getBuyer_number());
            tvStatus.setText(order.getStatus_text());
            tvSeller.setText("Продавец: " + order.getSeller());
            tvTimeAgo.setText(order.getTime_ago());

            // Set status color based on status
            switch (order.getStatus()) {
                case "red":
                    tvStatus.setBackgroundColor(Color.parseColor("#FF4444"));
                    break;
                case "brown":
                    tvStatus.setBackgroundColor(Color.parseColor("#8B4513"));
                    break;
                case "yellow":
                    tvStatus.setBackgroundColor(Color.parseColor("#FFFF44"));
                    tvStatus.setTextColor(Color.BLACK);
                    break;
                case "green":
                    tvStatus.setBackgroundColor(Color.parseColor("#44FF44"));
                    break;
                default:
                    tvStatus.setBackgroundColor(Color.GRAY);
                    break;
            }
        }
    }
}