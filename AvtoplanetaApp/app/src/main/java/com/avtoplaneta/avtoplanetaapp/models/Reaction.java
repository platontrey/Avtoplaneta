package com.avtoplaneta.avtoplanetaapp.models;

public class Reaction {
    private int id;
    private int messageId;
    private int userId;
    private String emoji;

    public Reaction() {
    }

    public Reaction(int id, int messageId, int userId, String emoji) {
        this.id = id;
        this.messageId = messageId;
        this.userId = userId;
        this.emoji = emoji;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public int getMessageId() {
        return messageId;
    }

    public void setMessageId(int messageId) {
        this.messageId = messageId;
    }

    public int getUserId() {
        return userId;
    }

    public void setUserId(int userId) {
        this.userId = userId;
    }

    public String getEmoji() {
        return emoji;
    }

    public void setEmoji(String emoji) {
        this.emoji = emoji;
    }
}