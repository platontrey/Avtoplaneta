package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class InventoryItem implements Serializable {
    private int id;
    private String name;
    private String description;
    private int quantity;
    private double price;
    private String category;
    private String salesman;
    private String location;
    private boolean status;
    private String brand;
    private String model;
    private String photo;
    private String inn;
    private String to_delete_at;
    private String to_delete_at_formatted;
    private String time_until_deletion;

    public InventoryItem() {
        // Default constructor for JSON deserialization
    }

    public InventoryItem(int id, String name, String description, int quantity, double price, String category) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.quantity = quantity;
        this.price = price;
        this.category = category;
    }

    // Геттеры и сеттеры
    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public int getQuantity() {
        return quantity;
    }

    public void setQuantity(int quantity) {
        this.quantity = quantity;
    }

    public double getPrice() {
        return price;
    }

    public void setPrice(double price) {
        this.price = price;
    }

    public String getCategory() {
        return category;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public String getSalesman() {
        return salesman;
    }

    public void setSalesman(String salesman) {
        this.salesman = salesman;
    }

    public String getLocation() {
        return location;
    }

    public void setLocation(String location) {
        this.location = location;
    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getBrand() {
        return brand;
    }

    public void setBrand(String brand) {
        this.brand = brand;
    }

    public String getModel() {
        return model;
    }

    public void setModel(String model) {
        this.model = model;
    }

    public String getPhoto() {
        return photo;
    }

    public void setPhoto(String photo) {
        this.photo = photo;
    }

    public String getInn() {
        return inn;
    }

    public void setInn(String inn) {
        this.inn = inn;
    }

    public String getTo_delete_at() {
        return to_delete_at;
    }

    public void setTo_delete_at(String to_delete_at) {
        this.to_delete_at = to_delete_at;
    }

    public String getTo_delete_at_formatted() {
        return to_delete_at_formatted;
    }

    public void setTo_delete_at_formatted(String to_delete_at_formatted) {
        this.to_delete_at_formatted = to_delete_at_formatted;
    }

    public String getTime_until_deletion() {
        return time_until_deletion;
    }

    public void setTime_until_deletion(String time_until_deletion) {
        this.time_until_deletion = time_until_deletion;
    }
}