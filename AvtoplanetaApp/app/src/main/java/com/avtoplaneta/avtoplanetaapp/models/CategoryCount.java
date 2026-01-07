package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;

public class CategoryCount implements Serializable {
    private String name;
    private int count;

    public CategoryCount() {
        // Default constructor for JSON deserialization
    }

    public CategoryCount(String name, int count) {
        this.name = name;
        this.count = count;
    }

    // Getters and setters
    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public int getCount() {
        return count;
    }

    public void setCount(int count) {
        this.count = count;
    }
}