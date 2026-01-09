package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class ConversationsResponse {
    private List<Conversation> conversations;

    public List<Conversation> getConversations() {
        return conversations;
    }

    public void setConversations(List<Conversation> conversations) {
        this.conversations = conversations;
    }
}