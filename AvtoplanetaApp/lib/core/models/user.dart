class User {
  final int id;
  final String email;
  final String name;
  final String initials;
  final String inn;
  final String provider;
  final String role;

  const User({
    required this.id,
    required this.email,
    required this.name,
    this.initials = '',
    this.inn = '',
    this.provider = 'local',
    required this.role,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] as int,
        email: json['email'] as String,
        name: json['name'] as String? ?? '',
        initials: json['initials'] as String? ?? '',
        inn: json['inn'] as String? ?? '',
        provider: json['provider'] as String? ?? 'local',
        role: json['role'] as String? ?? 'operator',
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'email': email,
        'name': name,
        'initials': initials,
        'inn': inn,
        'provider': provider,
        'role': role,
      };

  bool get isAdmin => role == 'admin';
  bool get isManager => role == 'manager' || role == 'admin';
  bool get isOperator => true; // все роли включают operator права
}

class LoginResponse {
  final User user;
  final String token;
  final String refreshToken;

  const LoginResponse({
    required this.user,
    required this.token,
    required this.refreshToken,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) => LoginResponse(
        user: User.fromJson(json['user'] as Map<String, dynamic>),
        token: json['token'] as String? ?? '',
        refreshToken: json['refresh_token'] as String? ?? '',
      );
}
