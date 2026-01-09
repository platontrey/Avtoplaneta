package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class Conversation {
    private int id;
    private List<Integer> participants;
    private String title;
    private String createdAt;
    private String lastMessageAt;
    private String lastMessage;
    private int unreadCount;

    public Conversation() {
    }

    public Conversation(int id, List<Integer> participants, String title, String createdAt, String lastMessageAt, String lastMessage, int unreadCount) {
        this.id = id;
        this.participants = participants;
        this.title = title;
        this.createdAt = createdAt;
        this.lastMessageAt = lastMessageAt;
        this.lastMessage = lastMessage;
        this.unreadCount = unreadCount;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public List<Integer> getParticipants() {
        return participants;
    }

    public void setParticipants(List<Integer> participants) {
        this.participants = participants;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    public String getLastMessageAt() {
        return lastMessageAt;
    }

    public void setLastMessageAt(String lastMessageAt) {
        this.lastMessageAt = lastMessageAt;
    }

    public String getLastMessage() {
        return lastMessage;
    }

    public void setLastMessage(String lastMessage) {
        this.lastMessage = lastMessage;
    }

    public int getUnreadCount() {
        return unreadCount;
    }

    public void setUnreadCount(int unreadCount) {
        this.unreadCount = unreadCount;
    }
}