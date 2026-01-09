package com.avtoplaneta.avtoplanetaapp.models;

public class User {
    private int id;
    private String email;
    private String name;
    private String initials;
    private String inn;
    private String provider;
    private String role;

    public User() {
    }

    public User(int id, String email, String name, String initials, String inn, String provider, String role) {
        this.id = id;
        this.email = email;
        this.name = name;
        this.initials = initials;
        this.inn = inn;
        this.provider = provider;
        this.role = role;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getInitials() {
        return initials;
    }

    public void setInitials(String initials) {
        this.initials = initials;
    }

    public String getInn() {
        return inn;
    }

    public void setInn(String inn) {
        this.inn = inn;
    }

    public String getProvider() {
        return provider;
    }

    public void setProvider(String provider) {
        this.provider = provider;
    }

    public String getRole() {
        return role;
    }

    public void setRole(String role) {
        this.role = role;
    }
}