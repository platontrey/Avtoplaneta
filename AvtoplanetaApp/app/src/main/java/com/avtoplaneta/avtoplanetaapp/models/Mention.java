package com.avtoplaneta.avtoplanetaapp.models;

public class Mention {
    private int id;
    private int messageId;
    private String type;
    private int resourceId;
    private String text;

    public Mention() {
    }

    public Mention(int id, int messageId, String type, int resourceId, String text) {
        this.id = id;
        this.messageId = messageId;
        this.type = type;
        this.resourceId = resourceId;
        this.text = text;
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

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public int getResourceId() {
        return resourceId;
    }

    public void setResourceId(int resourceId) {
        this.resourceId = resourceId;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }
}