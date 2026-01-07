package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class UpdateEarningsRequest implements Serializable {
    private double amount;

    public UpdateEarningsRequest() {
        // Default constructor for JSON deserialization
    }

    public UpdateEarningsRequest(double amount) {
        this.amount = amount;
    }

    // Getters and setters
    public double getAmount() {
        return amount;
    }

    public void setAmount(double amount) {
        this.amount = amount;
    }
}