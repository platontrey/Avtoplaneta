package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class MessagesResponse {
    private List<Message> messages;

    public List<Message> getMessages() {
        return messages;
    }

    public void setMessages(List<Message> messages) {
        this.messages = messages;
    }
}