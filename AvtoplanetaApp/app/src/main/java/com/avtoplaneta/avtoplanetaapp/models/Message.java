package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class Message {
    private int id;
    private int conversationId;
    private int senderId;
    private String content;
    private String messageType;
    private String voiceUrl;
    private String createdAt;
    private List<Integer> readBy;
    private List<String> attachments;
    private List<Mention> mentions;
    private List<Reaction> reactions;

    public Message() {
    }

    public Message(int id, int conversationId, int senderId, String content, String messageType, String voiceUrl, String createdAt, List<Integer> readBy, List<String> attachments, List<Mention> mentions, List<Reaction> reactions) {
        this.id = id;
        this.conversationId = conversationId;
        this.senderId = senderId;
        this.content = content;
        this.messageType = messageType;
        this.voiceUrl = voiceUrl;
        this.createdAt = createdAt;
        this.readBy = readBy;
        this.attachments = attachments;
        this.mentions = mentions;
        this.reactions = reactions;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public int getConversationId() {
        return conversationId;
    }

    public void setConversationId(int conversationId) {
        this.conversationId = conversationId;
    }

    public int getSenderId() {
        return senderId;
    }

    public void setSenderId(int senderId) {
        this.senderId = senderId;
    }

    public String getContent() {
        return content;
    }

    public void setContent(String content) {
        this.content = content;
    }

    public String getMessageType() {
        return messageType;
    }

    public void setMessageType(String messageType) {
        this.messageType = messageType;
    }

    public String getVoiceUrl() {
        return voiceUrl;
    }

    public void setVoiceUrl(String voiceUrl) {
        this.voiceUrl = voiceUrl;
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    public List<Integer> getReadBy() {
        return readBy;
    }

    public void setReadBy(List<Integer> readBy) {
        this.readBy = readBy;
    }

    public List<String> getAttachments() {
        return attachments;
    }

    public void setAttachments(List<String> attachments) {
        this.attachments = attachments;
    }

    public List<Mention> getMentions() {
        return mentions;
    }

    public void setMentions(List<Mention> mentions) {
        this.mentions = mentions;
    }

    public List<Reaction> getReactions() {
        return reactions;
    }

    public void setReactions(List<Reaction> reactions) {
        this.reactions = reactions;
    }
}