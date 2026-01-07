package com.avtoplaneta.avtoplanetaapp.models;

public class CsrfResponse {
    private String csrf_token;

    public CsrfResponse() {
    }

    public CsrfResponse(String csrf_token) {
        this.csrf_token = csrf_token;
    }

    public String getCsrfToken() {
        return csrf_token;
    }

    public void setCsrfToken(String csrf_token) {
        this.csrf_token = csrf_token;
    }
}