package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class AddOrderItemRequest implements Serializable {
    private int part_id;
    private int quantity;

    public AddOrderItemRequest() {
        // Default constructor for JSON deserialization
    }

    public AddOrderItemRequest(int part_id, int quantity) {
        this.part_id = part_id;
        this.quantity = quantity;
    }

    // Getters and setters
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