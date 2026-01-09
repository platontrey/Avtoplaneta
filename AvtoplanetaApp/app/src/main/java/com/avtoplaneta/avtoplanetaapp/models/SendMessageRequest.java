package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class SendMessageRequest {
    private String content;
    private List<String> attachments;

    public SendMessageRequest() {
    }

    public SendMessageRequest(String content, List<String> attachments) {
        this.content = content;
        this.attachments = attachments;
    }

    public String getContent() {
        return content;
    }

    public void setContent(String content) {
        this.content = content;
    }

    public List<String> getAttachments() {
        return attachments;
    }

    public void setAttachments(List<String> attachments) {
        this.attachments = attachments;
    }
}