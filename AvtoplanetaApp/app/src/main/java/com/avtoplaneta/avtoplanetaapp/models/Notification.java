package com.avtoplaneta.avtoplanetaapp.models;

public class Notification {
    private int id;
    private int userId;
    private String type;
    private String message;
    private boolean read;
    private String createdAt;
    private int relatedId;

    public Notification() {
    }

    public Notification(int id, int userId, String type, String message, boolean read, String createdAt, int relatedId) {
        this.id = id;
        this.userId = userId;
        this.type = type;
        this.message = message;
        this.read = read;
        this.createdAt = createdAt;
        this.relatedId = relatedId;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public int getUserId() {
        return userId;
    }

    public void setUserId(int userId) {
        this.userId = userId;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public boolean isRead() {
        return read;
    }

    public void setRead(boolean read) {
        this.read = read;
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    public int getRelatedId() {
        return relatedId;
    }

    public void setRelatedId(int relatedId) {
        this.relatedId = relatedId;
    }
}