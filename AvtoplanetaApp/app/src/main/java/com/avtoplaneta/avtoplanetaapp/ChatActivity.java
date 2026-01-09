package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.widget.EditText;
import android.widget.ImageButton;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.Message;
import com.avtoplaneta.avtoplanetaapp.models.MessagesResponse;
import com.avtoplaneta.avtoplanetaapp.models.SendMessageRequest;

import java.util.ArrayList;
import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class ChatActivity extends AppCompatActivity {

    private RecyclerView messagesRecyclerView;
    private MessagesAdapter messagesAdapter;
    private EditText messageEditText;
    private ImageButton sendButton;

    private int conversationId;
    private String conversationTitle;
    private List<Message> messages = new ArrayList<>();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_chat);

        // Получаем данные из Intent
        conversationId = getIntent().getIntExtra("conversation_id", -1);
        conversationTitle = getIntent().getStringExtra("conversation_title");

        if (conversationId == -1) {
            Toast.makeText(this, "Ошибка: ID чата не найден", Toast.LENGTH_SHORT).show();
            finish();
            return;
        }

        // Устанавливаем заголовок
        if (getSupportActionBar() != null) {
            getSupportActionBar().setTitle(conversationTitle != null ? conversationTitle : "Чат");
            getSupportActionBar().setDisplayHomeAsUpEnabled(true);
        }

        setupUI();
        loadMessages();
    }

    private void setupUI() {
        messagesRecyclerView = findViewById(R.id.messagesRecyclerView);
        messageEditText = findViewById(R.id.messageEditText);
        sendButton = findViewById(R.id.sendButton);

        // Настройка RecyclerView
        LinearLayoutManager layoutManager = new LinearLayoutManager(this);
        layoutManager.setStackFromEnd(true); // Показывать последние сообщения внизу
        messagesRecyclerView.setLayoutManager(layoutManager);
        messagesAdapter = new MessagesAdapter(messages);
        messagesRecyclerView.setAdapter(messagesAdapter);

        // Обработчик отправки сообщения
        sendButton.setOnClickListener(v -> sendMessage());
    }

    private void loadMessages() {
        Log.d("ChatActivity", "Loading messages for conversation: " + conversationId);

        ApiService apiService = RetrofitClient.getApiService();
        Call<MessagesResponse> call = apiService.getMessages(conversationId);

        call.enqueue(new Callback<MessagesResponse>() {
            @Override
            public void onResponse(Call<MessagesResponse> call, Response<MessagesResponse> response) {
                if (response.isSuccessful() && response.body() != null) {
                    messages.clear();
                    messages.addAll(response.body().getMessages());
                    messagesAdapter.notifyDataSetChanged();
                    scrollToBottom();
                    Log.d("ChatActivity", "Loaded " + messages.size() + " messages");
                } else {
                    Log.e("ChatActivity", "Error loading messages: " + response.code());
                    Toast.makeText(ChatActivity.this, "Ошибка загрузки сообщений", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<MessagesResponse> call, Throwable t) {
                Log.e("ChatActivity", "Network error: " + t.getMessage());
                Toast.makeText(ChatActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void sendMessage() {
        String content = messageEditText.getText().toString().trim();
        if (content.isEmpty()) {
            return;
        }

        SendMessageRequest request = new SendMessageRequest(content, null);

        ApiService apiService = RetrofitClient.getApiService();
        Call<Message> call = apiService.sendMessage(conversationId, request);

        call.enqueue(new Callback<Message>() {
            @Override
            public void onResponse(Call<Message> call, Response<Message> response) {
                if (response.isSuccessful() && response.body() != null) {
                    messages.add(response.body());
                    messagesAdapter.notifyItemInserted(messages.size() - 1);
                    scrollToBottom();
                    messageEditText.setText("");
                    Log.d("ChatActivity", "Message sent successfully");
                } else {
                    Log.e("ChatActivity", "Error sending message: " + response.code());
                    Toast.makeText(ChatActivity.this, "Ошибка отправки сообщения", Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<Message> call, Throwable t) {
                Log.e("ChatActivity", "Network error: " + t.getMessage());
                Toast.makeText(ChatActivity.this, "Ошибка сети", Toast.LENGTH_SHORT).show();
            }
        });
    }

    private void scrollToBottom() {
        if (messagesAdapter.getItemCount() > 0) {
            messagesRecyclerView.smoothScrollToPosition(messagesAdapter.getItemCount() - 1);
        }
    }

    @Override
    public boolean onSupportNavigateUp() {
        finish();
        return true;
    }
}