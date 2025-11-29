package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class OrderItem implements Serializable {
    private int id;
    private int order_id;
    private int part_id;
    private int quantity;

    public OrderItem() {
        // Default constructor for JSON deserialization
    }

    // Getters and setters
    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public int getOrder_id() {
        return order_id;
    }

    public void setOrder_id(int order_id) {
        this.order_id = order_id;
    }

    public int getPart_id() {
        return part_id;
    }

    public void setPart_id(int part_id) {
        this.part_id = part_id;
    }

    public int getQuantity() {
        return quantity;
    }

    public void setQuantity(int quantity) {
        this.quantity = quantity;
    }
}