package com.avtoplaneta.avtoplanetaapp.api;

import com.avtoplaneta.avtoplanetaapp.models.AddOrderItemRequest;
import com.avtoplaneta.avtoplanetaapp.models.BulkUpdateRequest;
import com.avtoplaneta.avtoplanetaapp.models.CsrfResponse;
import com.avtoplaneta.avtoplanetaapp.models.InventoryItem;
import com.avtoplaneta.avtoplanetaapp.models.LoginRequest;
import com.avtoplaneta.avtoplanetaapp.models.LoginResponse;
import com.avtoplaneta.avtoplanetaapp.models.MonthlySales;
import com.avtoplaneta.avtoplanetaapp.models.Order;
import com.avtoplaneta.avtoplanetaapp.models.StatisticsResponse;
import com.avtoplaneta.avtoplanetaapp.models.UpdateEarningsRequest;

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
}