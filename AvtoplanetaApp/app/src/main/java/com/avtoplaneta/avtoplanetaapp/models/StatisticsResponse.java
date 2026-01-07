package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;
import java.util.List;

public class StatisticsResponse implements Serializable {
    private int total_parts;
    private int total_quantity;
    private double total_value;
    private double total_earnings;
    private List<CategoryCount> categories;
    private List<MonthlySales> monthly_sales;

    public StatisticsResponse() {
        // Default constructor for JSON deserialization
    }

    // Getters and setters
    public int getTotal_parts() {
        return total_parts;
    }

    public void setTotal_parts(int total_parts) {
        this.total_parts = total_parts;
    }

    public int getTotal_quantity() {
        return total_quantity;
    }

    public void setTotal_quantity(int total_quantity) {
        this.total_quantity = total_quantity;
    }

    public double getTotal_value() {
        return total_value;
    }

    public void setTotal_value(double total_value) {
        this.total_value = total_value;
    }

    public double getTotal_earnings() {
        return total_earnings;
    }

    public void setTotal_earnings(double total_earnings) {
        this.total_earnings = total_earnings;
    }

    public List<CategoryCount> getCategories() {
        return categories;
    }

    public void setCategories(List<CategoryCount> categories) {
        this.categories = categories;
    }

    public List<MonthlySales> getMonthly_sales() {
        return monthly_sales;
    }

    public void setMonthly_sales(List<MonthlySales> monthly_sales) {
        this.monthly_sales = monthly_sales;
    }
}