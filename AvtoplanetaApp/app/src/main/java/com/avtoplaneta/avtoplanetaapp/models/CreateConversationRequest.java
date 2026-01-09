package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class CreateConversationRequest {
    private List<Integer> participants;
    private String title;

    public CreateConversationRequest() {
    }

    public CreateConversationRequest(List<Integer> participants, String title) {
        this.participants = participants;
        this.title = title;
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
}