package com.avtoplaneta.avtoplanetaapp.models;

import java.util.List;

public class NotificationsResponse {
    private List<Notification> notifications;

    public List<Notification> getNotifications() {
        return notifications;
    }

    public void setNotifications(List<Notification> notifications) {
        this.notifications = notifications;
    }
}