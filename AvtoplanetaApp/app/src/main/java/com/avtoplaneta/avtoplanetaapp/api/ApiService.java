package com.avtoplaneta.avtoplanetaapp.api;

import com.avtoplaneta.avtoplanetaapp.models.AddOrderItemRequest;
import com.avtoplaneta.avtoplanetaapp.models.AddReactionRequest;
import com.avtoplaneta.avtoplanetaapp.models.BulkUpdateRequest;
import com.avtoplaneta.avtoplanetaapp.models.Car;
import com.avtoplaneta.avtoplanetaapp.models.Conversation;
import com.avtoplaneta.avtoplanetaapp.models.ConversationsResponse;
import com.avtoplaneta.avtoplanetaapp.models.CreateConversationRequest;
import com.avtoplaneta.avtoplanetaapp.models.CsrfResponse;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.avtoplaneta.avtoplanetaapp.models.LoginRequest;
import com.avtoplaneta.avtoplanetaapp.models.LoginResponse;
import com.avtoplaneta.avtoplanetaapp.models.Message;
import com.avtoplaneta.avtoplanetaapp.models.MessagesResponse;
import com.avtoplaneta.avtoplanetaapp.models.MonthlySales;
import com.avtoplaneta.avtoplanetaapp.models.Notification;
import com.avtoplaneta.avtoplanetaapp.models.NotificationsResponse;
import com.avtoplaneta.avtoplanetaapp.models.Order;
import com.avtoplaneta.avtoplanetaapp.models.Reaction;
import com.avtoplaneta.avtoplanetaapp.models.SendMessageRequest;
import com.avtoplaneta.avtoplanetaapp.models.StatisticsResponse;
import com.avtoplaneta.avtoplanetaapp.models.UpdateEarningsRequest;
import com.avtoplaneta.avtoplanetaapp.models.User;
import com.avtoplaneta.avtoplanetaapp.models.UsersResponse;

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

    @GET("auth/csrf-token")
    Call<CsrfResponse> getCsrfToken();

    @GET("api/inventory")
    Call<List<InventoryItem>> getInventory(@retrofit2.http.Query("search") String search,
                                           @retrofit2.http.Query("category") String category,
                                           @retrofit2.http.Query("brand") String brand,
                                           @retrofit2.http.Query("model") String model,
                                           @retrofit2.http.Query("location") String location,
                                           @retrofit2.http.Query("status") String status,
                                           @retrofit2.http.Query("hasPhoto") String hasPhoto,
                                           @retrofit2.http.Query("limit") Integer limit,
                                           @retrofit2.http.Query("page") Integer page);

    @GET("api/inventory")
    Call<List<InventoryItem>> getInventory();

    @POST("api/addpart")
    Call<InventoryItem> addPart(@Body InventoryItem item);

    @PUT("api/updatepart/{id}")
    Call<InventoryItem> updatePart(@Path("id") int id, @Body InventoryItem item);

    @Multipart
    @POST("api/uploadpartphoto/{id}")
    Call<Object> uploadPartPhoto(@Path("id") int id, @Part MultipartBody.Part photo);

    @DELETE("api/deletepart/{id}")
    Call<Object> deletePart(@Path("id") int id);

    @DELETE("api/deletepartphoto/{id}")
    Call<Object> deletePartPhoto(@Path("id") int id);

    @POST("api/markpartfordeletion/{id}")
    Call<Object> markPartForDeletion(@Path("id") int id);

    @GET("api/statistics")
    Call<StatisticsResponse> getStatistics();

    @POST("api/statistics/update-earnings")
    Call<Object> updateEarnings(@Body UpdateEarningsRequest request);

    // Orders API methods
    @GET("orders")
    Call<List<Order>> getOrders();

    @POST("orders")
    Call<Order> createOrder(@Body Order order);

    @PUT("admin/orders/{id}/status")
    Call<Order> updateOrderStatus(@Path("id") int id, @Body Order order);

    @DELETE("admin/orders/{id}")
    Call<Object> deleteOrder(@Path("id") int id);

    @PUT("admin/orders/{id}/complete")
    Call<Object> completeOrder(@Path("id") int id);

    @POST("orders/{id}/items")
    Call<Object> addOrderItem(@Path("id") int id, @Body AddOrderItemRequest request);

    @GET("monthly-sales")
    Call<List<MonthlySales>> getMonthlySales();

    @GET("api/export/xml")
    Call<Object> exportXML();

    @POST("api/export/drom")
    Call<Object> exportDrom();

    @POST("api/defect-reports")
    Call<Object> createDefectReport(@Body Object request);

    @DELETE("api/admin/delete-zero-quantity-parts/{supplier_code}")
    Call<Object> deleteZeroQuantityParts(@Path("supplier_code") String supplierCode);

    @GET("api/admin/supplier-codes")
    Call<List<String>> getSupplierCodes();

    @DELETE("api/admin/bulk-delete-parts")
    Call<Object> bulkDeleteParts(@Body List<Integer> ids);

    @PUT("api/admin/bulk-update-parts")
    Call<Object> bulkUpdateParts(@Body BulkUpdateRequest request);

    // Cars API methods
    @GET("api/cars")
    Call<List<Car>> getCars();

    @POST("api/cars")
    Call<Car> addCar(@Body Car car);

    @PUT("api/cars/{id}")
    Call<Car> updateCar(@Path("id") int id, @Body Car car);

    @DELETE("api/cars/{id}")
    Call<Object> deleteCar(@Path("id") int id);

    // Messaging API methods
    @GET("api/messaging/conversations")
    Call<ConversationsResponse> getConversations();

    @POST("api/messaging/conversations")
    Call<Conversation> createConversation(@Body CreateConversationRequest request);

    @GET("api/messaging/conversations/{id}")
    Call<Conversation> getConversation(@Path("id") int id);

    @PUT("api/messaging/conversations/{id}")
    Call<Conversation> updateConversation(@Path("id") int id, @Body Conversation conversation);

    @DELETE("api/messaging/conversations/{id}")
    Call<Object> deleteConversation(@Path("id") int id);

    @DELETE("api/messaging/conversations/{id}/participants/{userId}")
    Call<Object> removeParticipant(@Path("id") int conversationId, @Path("userId") int userId);

    @GET("api/messaging/conversations/{id}/messages")
    Call<MessagesResponse> getMessages(@Path("id") int conversationId);

    @POST("api/messaging/conversations/{id}/messages")
    Call<Message> sendMessage(@Path("id") int conversationId, @Body SendMessageRequest request);

    @Multipart
    @POST("api/messaging/conversations/{id}/messages/voice")
    Call<Message> sendVoiceMessage(@Path("id") int conversationId, @Part MultipartBody.Part voice);

    @DELETE("api/messaging/messages/{id}")
    Call<Object> deleteMessage(@Path("id") int messageId);

    @PUT("api/messaging/messages/{id}/read")
    Call<Object> markMessageRead(@Path("id") int messageId);

    @POST("api/messaging/messages/{id}/reactions")
    Call<Reaction> addReaction(@Path("id") int messageId, @Body AddReactionRequest request);

    @DELETE("api/messaging/messages/{id}/reactions/{reactionId}")
    Call<Object> removeReaction(@Path("id") int messageId, @Path("reactionId") int reactionId);

    @GET("api/messaging/notifications")
    Call<NotificationsResponse> getNotifications();

    @PUT("api/messaging/notifications/{id}/read")
    Call<Object> markNotificationRead(@Path("id") int notificationId);

    @PUT("api/messaging/notifications/read-all")
    Call<Object> markAllNotificationsRead();

    @GET("api/messaging/users")
    Call<UsersResponse> getUsers();

    @GET("api/messaging/search")
    Call<MessagesResponse> searchMessages(@retrofit2.http.Query("q") String query);
}