package com.example;

import java.sql.*;
import java.util.ArrayList;
import java.util.List;

public class UserService {

    private final Connection conn;

    public UserService(Connection conn) {
        this.conn = conn;
    }

    // BUG: SQL injection via string concatenation
    public User findByName(String name) throws SQLException {
        String sql = "SELECT * FROM users WHERE name = '" + name + "'";
        Statement stmt = conn.createStatement();
        ResultSet rs = stmt.executeQuery(sql);
        if (rs.next()) {
            return new User(rs.getInt("id"), rs.getString("name"), rs.getString("email"));
        }
        return null;
    }

    // BUG: SQL injection via String.format
    public List<User> findByRole(String role) throws SQLException {
        String sql = String.format("SELECT * FROM users WHERE role = '%s' AND active = 1", role);
        Statement stmt = conn.createStatement();
        ResultSet rs = stmt.executeQuery(sql);
        List<User> users = new ArrayList<>();
        while (rs.next()) {
            users.add(new User(rs.getInt("id"), rs.getString("name"), rs.getString("email")));
        }
        return users;
    }

    // BUG: SQL injection via ORDER BY concat
    public List<User> findAllSorted(String sortBy) throws SQLException {
        String sql = "SELECT * FROM users ORDER BY " + sortBy;
        Statement stmt = conn.createStatement();
        ResultSet rs = stmt.executeQuery(sql);
        List<User> users = new ArrayList<>();
        while (rs.next()) {
            users.add(new User(rs.getInt("id"), rs.getString("name"), rs.getString("email")));
        }
        return users;
    }

    // BUG: Hardcoded password
    private static final String DB_PASSWORD = "MySuperSecret123!";
    private static final String API_KEY = "sk-proj-abcdefghijklmnopqrstuvwxyz123456";

    // GOOD: Parameterized query (should NOT trigger SQL injection)
    public User findById(int id) throws SQLException {
        PreparedStatement ps = conn.prepareStatement("SELECT * FROM users WHERE id = ?");
        ps.setInt(1, id);
        ResultSet rs = ps.executeQuery();
        if (rs.next()) {
            return new User(rs.getInt("id"), rs.getString("name"), rs.getString("email"));
        }
        return null;
    }

    // BUG: Empty catch block
    public void updateEmail(int id, String email) {
        try {
            String sql = "UPDATE users SET email = " + email + " WHERE id = " + id;
            conn.createStatement().executeUpdate(sql);
        } catch (Exception e) {
            // ignore
        }
    }
}
