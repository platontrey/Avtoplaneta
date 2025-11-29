package com.avtoplaneta.avtoplanetaapp.models;

public class LoginResponse {
    private User user;
    private String message;

    public LoginResponse() {
    }

    public LoginResponse(User user, String message) {
        this.user = user;
        this.message = message;
    }

    public User getUser() {
        return user;
    }

    public void setUser(User user) {
        this.user = user;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }
}