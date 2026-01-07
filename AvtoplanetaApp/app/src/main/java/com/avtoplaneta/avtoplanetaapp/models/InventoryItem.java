package com.avtoplaneta.avtoplanetaapp.models;

import java.io.Serializable;
import java.util.List;

public class InventoryItem implements Serializable {
    private int id;
    private String name;
    private String description;
    private int quantity;
    private double price;
    private String category;
    private String salesman;
    private String location;
    private boolean status;
    private String brand;
    private String model;
    private String photo;
    private List<String> photos;
    private String inn;
    private String vin;
    private String to_delete_at;
    private String to_delete_at_formatted;
    private String time_until_deletion;

    // Характеристики запчасти
    private String body_brand;
    private String engine_brand;
    private String car_release_date;
    private String front_rear;
    private String left_right;
    private String top_bottom;
    private String number;
    private String manufacturer;
    private String manufacturer_code;
    private String oem_code;
    private String color;
    private String condition;
    private String supplier_code;
    private String defect;
    private String transmission;
    private String drive;
    private String wear_percentage;
    private String season;
    private String diameter;
    private String width;
    private String profile;
    private String tire_quantity;
    private String drilling;
    private String offset;
    private String center_hole_diameter;
    private String tire_model;

    public InventoryItem() {
        // Default constructor for JSON deserialization
    }

    public InventoryItem(int id, String name, String description, int quantity, double price, String category) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.quantity = quantity;
        this.price = price;
        this.category = category;
    }

    // Геттеры и сеттеры
    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public int getQuantity() {
        return quantity;
    }

    public void setQuantity(int quantity) {
        this.quantity = quantity;
    }

    public double getPrice() {
        return price;
    }

    public void setPrice(double price) {
        this.price = price;
    }

    public String getCategory() {
        return category;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public String getSalesman() {
        return salesman;
    }

    public void setSalesman(String salesman) {
        this.salesman = salesman;
    }

    public String getLocation() {
        return location;
    }

    public void setLocation(String location) {
        this.location = location;
    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getBrand() {
        return brand;
    }

    public void setBrand(String brand) {
        this.brand = brand;
    }

    public String getModel() {
        return model;
    }

    public void setModel(String model) {
        this.model = model;
    }

    public String getPhoto() {
        return photo;
    }

    public void setPhoto(String photo) {
        this.photo = photo;
    }

    public List<String> getPhotos() {
        return photos;
    }

    public void setPhotos(List<String> photos) {
        this.photos = photos;
    }

    public String getInn() {
        return inn;
    }

    public void setInn(String inn) {
        this.inn = inn;
    }

    public String getVin() {
        return vin;
    }

    public void setVin(String vin) {
        this.vin = vin;
    }

    public String getTo_delete_at() {
        return to_delete_at;
    }

    public void setTo_delete_at(String to_delete_at) {
        this.to_delete_at = to_delete_at;
    }

    public String getTo_delete_at_formatted() {
        return to_delete_at_formatted;
    }

    public void setTo_delete_at_formatted(String to_delete_at_formatted) {
        this.to_delete_at_formatted = to_delete_at_formatted;
    }

    public String getTime_until_deletion() {
        return time_until_deletion;
    }

    public void setTime_until_deletion(String time_until_deletion) {
        this.time_until_deletion = time_until_deletion;
    }

    // Геттеры и сеттеры для характеристик запчасти
    public String getBody_brand() {
        return body_brand;
    }

    public void setBody_brand(String body_brand) {
        this.body_brand = body_brand;
    }

    public String getEngine_brand() {
        return engine_brand;
    }

    public void setEngine_brand(String engine_brand) {
        this.engine_brand = engine_brand;
    }

    public String getCar_release_date() {
        return car_release_date;
    }

    public void setCar_release_date(String car_release_date) {
        this.car_release_date = car_release_date;
    }

    public String getFront_rear() {
        return front_rear;
    }

    public void setFront_rear(String front_rear) {
        this.front_rear = front_rear;
    }

    public String getLeft_right() {
        return left_right;
    }

    public void setLeft_right(String left_right) {
        this.left_right = left_right;
    }

    public String getTop_bottom() {
        return top_bottom;
    }

    public void setTop_bottom(String top_bottom) {
        this.top_bottom = top_bottom;
    }

    public String getNumber() {
        return number;
    }

    public void setNumber(String number) {
        this.number = number;
    }

    public String getManufacturer() {
        return manufacturer;
    }

    public void setManufacturer(String manufacturer) {
        this.manufacturer = manufacturer;
    }

    public String getManufacturer_code() {
        return manufacturer_code;
    }

    public void setManufacturer_code(String manufacturer_code) {
        this.manufacturer_code = manufacturer_code;
    }

    public String getOem_code() {
        return oem_code;
    }

    public void setOem_code(String oem_code) {
        this.oem_code = oem_code;
    }

    public String getColor() {
        return color;
    }

    public void setColor(String color) {
        this.color = color;
    }

    public String getCondition() {
        return condition;
    }

    public void setCondition(String condition) {
        this.condition = condition;
    }

    public String getSupplier_code() {
        return supplier_code;
    }

    public void setSupplier_code(String supplier_code) {
        this.supplier_code = supplier_code;
    }

    public String getDefect() {
        return defect;
    }

    public void setDefect(String defect) {
        this.defect = defect;
    }

    public String getTransmission() {
        return transmission;
    }

    public void setTransmission(String transmission) {
        this.transmission = transmission;
    }

    public String getDrive() {
        return drive;
    }

    public void setDrive(String drive) {
        this.drive = drive;
    }

    public String getWear_percentage() {
        return wear_percentage;
    }

    public void setWear_percentage(String wear_percentage) {
        this.wear_percentage = wear_percentage;
    }

    public String getSeason() {
        return season;
    }

    public void setSeason(String season) {
        this.season = season;
    }

    public String getDiameter() {
        return diameter;
    }

    public void setDiameter(String diameter) {
        this.diameter = diameter;
    }

    public String getWidth() {
        return width;
    }

    public void setWidth(String width) {
        this.width = width;
    }

    public String getProfile() {
        return profile;
    }

    public void setProfile(String profile) {
        this.profile = profile;
    }

    public String getTire_quantity() {
        return tire_quantity;
    }

    public void setTire_quantity(String tire_quantity) {
        this.tire_quantity = tire_quantity;
    }

    public String getDrilling() {
        return drilling;
    }

    public void setDrilling(String drilling) {
        this.drilling = drilling;
    }

    public String getOffset() {
        return offset;
    }

    public void setOffset(String offset) {
        this.offset = offset;
    }

    public String getCenter_hole_diameter() {
        return center_hole_diameter;
    }

    public void setCenter_hole_diameter(String center_hole_diameter) {
        this.center_hole_diameter = center_hole_diameter;
    }

    public String getTire_model() {
        return tire_model;
    }

    public void setTire_model(String tire_model) {
        this.tire_model = tire_model;
    }
}