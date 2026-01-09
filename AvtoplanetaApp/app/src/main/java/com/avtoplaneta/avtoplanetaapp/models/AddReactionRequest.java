package com.avtoplaneta.avtoplanetaapp.models;

public class AddReactionRequest {
    private String emoji;

    public AddReactionRequest() {
    }

    public AddReactionRequest(String emoji) {
        this.emoji = emoji;
    }

    public String getEmoji() {
        return emoji;
    }

    public void setEmoji(String emoji) {
        this.emoji = emoji;
    }
}