package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class MonthlySales implements Serializable {
    private String month;
    private double sales;

    public MonthlySales() {
        // Default constructor for JSON deserialization
    }

    public MonthlySales(String month, double sales) {
        this.month = month;
        this.sales = sales;
    }

    // Getters and setters
    public String getMonth() {
        return month;
    }

    public void setMonth(String month) {
        this.month = month;
    }

    public double getSales() {
        return sales;
    }

    public void setSales(double sales) {
        this.sales = sales;
    }
}