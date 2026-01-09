package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;
import java.util.List;

public class Order implements Serializable {
    private int id;
    private int customer_id;
    private int seller_id;
    private String seller;
    private String part;
    private int part_id;
    private int car_id;
    private String order_number;
    private String buyer_number;
    private String status;
    private String status_text;
    private String created_at;
    private String created_at_formatted;
    private String time_ago;
    private List<OrderItem> items;

    public Order() {
        // Default constructor for JSON deserialization
    }

    // Getters and setters
    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public int getCustomer_id() {
        return customer_id;
    }

    public void setCustomer_id(int customer_id) {
        this.customer_id = customer_id;
    }

    public int getSeller_id() {
        return seller_id;
    }

    public void setSeller_id(int seller_id) {
        this.seller_id = seller_id;
    }

    public String getSeller() {
        return seller;
    }

    public void setSeller(String seller) {
        this.seller = seller;
    }

    public String getPart() {
        return part;
    }

    public void setPart(String part) {
        this.part = part;
    }

    public int getPart_id() {
        return part_id;
    }

    public void setPart_id(int part_id) {
        this.part_id = part_id;
    }

    public int getCar_id() {
        return car_id;
    }

    public void setCar_id(int car_id) {
        this.car_id = car_id;
    }

    public String getOrder_number() {
        return order_number;
    }

    public void setOrder_number(String order_number) {
        this.order_number = order_number;
    }

    public String getBuyer_number() {
        return buyer_number;
    }

    public void setBuyer_number(String buyer_number) {
        this.buyer_number = buyer_number;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getStatus_text() {
        return status_text;
    }

    public void setStatus_text(String status_text) {
        this.status_text = status_text;
    }

    public String getCreated_at() {
        return created_at;
    }

    public void setCreated_at(String created_at) {
        this.created_at = created_at;
    }

    public String getCreated_at_formatted() {
        return created_at_formatted;
    }

    public void setCreated_at_formatted(String created_at_formatted) {
        this.created_at_formatted = created_at_formatted;
    }

    public String getTime_ago() {
        return time_ago;
    }

    public void setTime_ago(String time_ago) {
        this.time_ago = time_ago;
    }

    public List<OrderItem> getItems() {
        return items;
    }

    public void setItems(List<OrderItem> items) {
        this.items = items;
    }
}