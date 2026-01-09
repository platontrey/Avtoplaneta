package com.avtoplaneta.avtoplanetaapp;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.avtoplaneta.avtoplanetaapp.api.ApiService;
import com.avtoplaneta.avtoplanetaapp.api.RetrofitClient;
import com.avtoplaneta.avtoplanetaapp.models.Conversation;
import com.avtoplaneta.avtoplanetaapp.models.ConversationsResponse;

import java.util.ArrayList;
import java.util.List;

import retrofit2.Call;
import retrofit2.Callback;
import retrofit2.Response;

public class ConversationsFragment extends Fragment implements ConversationsAdapter.OnConversationClickListener {

    private List<Conversation> conversations;
    private ConversationsAdapter conversationsAdapter;

    @Nullable
    @Override
    public View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container, @Nullable Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.fragment_conversations, container, false);

        // Инициализация данных
        conversations = new ArrayList<>();

        // Настройка адаптеров
        setupAdapters(view);

        // Загрузка данных
        loadConversations();

        return view;
    }

    private void setupAdapters(View view) {
        conversationsAdapter = new ConversationsAdapter(conversations, this);
        RecyclerView conversationsRecyclerView = view.findViewById(R.id.conversationsRecyclerView);
        conversationsRecyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        conversationsRecyclerView.setAdapter(conversationsAdapter);
    }

    private void loadConversations() {
        Log.d("ConversationsFragment", "Loading conversations");

        ApiService apiService = RetrofitClient.getApiService();
        Call<ConversationsResponse> call = apiService.getConversations();

        call.enqueue(new Callback<ConversationsResponse>() {
            @Override
            public void onResponse(Call<ConversationsResponse> call, Response<ConversationsResponse> response) {
                if (response.isSuccessful() && response.body() != null) {
                    conversations.clear();
                    conversations.addAll(response.body().getConversations());
                    conversationsAdapter.notifyDataSetChanged();
                    Log.d("ConversationsFragment", "Conversations loaded: " + conversations.size());
                } else {
                    Log.e("ConversationsFragment", "Error loading conversations: " + response.code());
                    Toast.makeText(getContext(), "Ошибка загрузки чатов: " + response.code(), Toast.LENGTH_SHORT).show();
                }
            }

            @Override
            public void onFailure(Call<ConversationsResponse> call, Throwable t) {
                Log.e("ConversationsFragment", "Network error: " + t.getMessage());
                Toast.makeText(getContext(), "Ошибка сети: " + t.getMessage(), Toast.LENGTH_SHORT).show();
            }
        });
    }

    @Override
    public void onConversationClick(Conversation conversation) {
        // Открываем чат
        Intent intent = new Intent(getContext(), ChatActivity.class);
        intent.putExtra("conversation_id", conversation.getId());
        intent.putExtra("conversation_title", conversation.getTitle());
        startActivity(intent);
    }
}