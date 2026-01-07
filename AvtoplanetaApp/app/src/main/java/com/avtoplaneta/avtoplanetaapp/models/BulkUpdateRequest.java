package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;
import java.util.List;

public class BulkUpdateRequest implements Serializable {
    private List<Integer> ids;
    private BulkUpdateFields updates;

    public BulkUpdateRequest() {
        // Default constructor for JSON deserialization
    }

    public BulkUpdateRequest(List<Integer> ids, BulkUpdateFields updates) {
        this.ids = ids;
        this.updates = updates;
    }

    // Getters and setters
    public List<Integer> getIds() {
        return ids;
    }

    public void setIds(List<Integer> ids) {
        this.ids = ids;
    }

    public BulkUpdateFields getUpdates() {
        return updates;
    }

    public void setUpdates(BulkUpdateFields updates) {
        this.updates = updates;
    }

    public static class BulkUpdateFields implements Serializable {
        private String category;
        private String brand;
        private String model;
        private String location;
        private Double price;
        private Integer quantity;
        private String description;

        public BulkUpdateFields() {
            // Default constructor
        }

        // Getters and setters
        public String getCategory() {
            return category;
        }

        public void setCategory(String category) {
            this.category = category;
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

        public String getLocation() {
            return location;
        }

        public void setLocation(String location) {
            this.location = location;
        }

        public Double getPrice() {
            return price;
        }

        public void setPrice(Double price) {
            this.price = price;
        }

        public Integer getQuantity() {
            return quantity;
        }

        public void setQuantity(Integer quantity) {
            this.quantity = quantity;
        }

        public String getDescription() {
            return description;
        }

        public void setDescription(String description) {
            this.description = description;
        }
    }
}