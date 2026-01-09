package com.avtoplaneta.avtoplanetaapp;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.models.Conversation;

import java.util.List;

public class ConversationsAdapter extends RecyclerView.Adapter<ConversationsAdapter.ConversationViewHolder> {

    private List<Conversation> conversations;
    private OnConversationClickListener listener;

    public interface OnConversationClickListener {
        void onConversationClick(Conversation conversation);
    }

    public ConversationsAdapter(List<Conversation> conversations, OnConversationClickListener listener) {
        this.conversations = conversations;
        this.listener = listener;
    }

    @NonNull
    @Override
    public ConversationViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_conversation, parent, false);
        return new ConversationViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull ConversationViewHolder holder, int position) {
        Conversation conversation = conversations.get(position);
        holder.bind(conversation, listener);
    }

    @Override
    public int getItemCount() {
        return conversations.size();
    }

    static class ConversationViewHolder extends RecyclerView.ViewHolder {
        private TextView tvConversationTitle;
        private TextView tvLastMessage;
        private TextView tvUnreadCount;

        public ConversationViewHolder(@NonNull View itemView) {
            super(itemView);
            tvConversationTitle = itemView.findViewById(R.id.tvConversationTitle);
            tvLastMessage = itemView.findViewById(R.id.tvLastMessage);
            tvUnreadCount = itemView.findViewById(R.id.tvUnreadCount);
        }

        public void bind(Conversation conversation, OnConversationClickListener listener) {
            String title = conversation.getTitle();
            if (title == null || title.isEmpty()) {
                title = "Чат #" + conversation.getId();
            }
            tvConversationTitle.setText(title);

            String lastMessage = conversation.getLastMessage();
            if (lastMessage == null || lastMessage.isEmpty()) {
                lastMessage = "Нет сообщений";
            }
            tvLastMessage.setText(lastMessage);

            int unreadCount = conversation.getUnreadCount();
            if (unreadCount > 0) {
                tvUnreadCount.setVisibility(View.VISIBLE);
                tvUnreadCount.setText(String.valueOf(unreadCount));
            } else {
                tvUnreadCount.setVisibility(View.GONE);
            }

            itemView.setOnClickListener(v -> {
                if (listener != null) {
                    listener.onConversationClick(conversation);
                }
            });
        }
    }
}