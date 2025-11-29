package com.avtoplaneta.avtoplanetaapp.api;

import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.avtoplaneta.avtoplanetaapp.models.LoginRequest;
import com.avtoplaneta.avtoplanetaapp.models.LoginResponse;
import com.avtoplaneta.avtoplanetaapp.models.Order;

import java.util.List;

import okhttp3.MultipartBody;
import retrofit2.Call;
import retrofit2.http.Body;
import retrofit2.http.DELETE;
import retrofit2.http.GET;
import retrofit2.http.Multipart;
import retrofit2.http.POST;
import retrofit2.http.PUT;
import retrofit2.http.Part;
import retrofit2.http.Path;

public interface ApiService {
    @POST("auth/login")
    Call<LoginResponse> login(@Body LoginRequest loginRequest);

    @GET("api/inventory")
    Call<List<InventoryItem>> getInventory();

    @POST("api/addpart")
    Call<InventoryItem> addPart(@Body InventoryItem item);

    @PUT("api/updatepart/{id}")
    Call<InventoryItem> updatePart(@Path("id") int id, @Body InventoryItem item);

    @Multipart
    @POST("api/uploadpartphoto/{id}")
    Call<Object> uploadPartPhoto(@Path("id") int id, @Part MultipartBody.Part photo);

    // Orders API methods
    @GET("orders")
    Call<List<Order>> getOrders();

    @POST("orders")
    Call<Order> createOrder(@Body Order order);

    @PUT("admin/orders/{id}/status")
    Call<Order> updateOrderStatus(@Path("id") int id, @Body Order order);

    @DELETE("admin/orders/{id}")
    Call<Object> deleteOrder(@Path("id") int id);
}