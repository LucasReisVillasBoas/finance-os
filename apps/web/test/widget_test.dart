import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:financeos_web/features/auth/screens/onboarding_screen.dart';
import 'package:financeos_web/features/auth/screens/login_screen.dart';
import 'package:financeos_web/features/auth/screens/register_screen.dart';

void main() {
  testWidgets('OnboardingScreen renders slides and navigation buttons', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: OnboardingScreen(),
        ),
      ),
    );

    expect(find.text('Pular'), findsOneWidget);
    expect(find.text('Próximo'), findsOneWidget);
  });

  testWidgets('LoginScreen renders email and password fields', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: LoginScreen(),
        ),
      ),
    );

    expect(find.text('E-mail'), findsOneWidget);
    expect(find.text('Senha'), findsOneWidget);
    expect(find.text('Entrar'), findsWidgets);
  });

  testWidgets('RegisterScreen shows validation errors on empty submit', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: RegisterScreen(),
        ),
      ),
    );

    // Tap the register button without filling in the form
    final registerButton = find.widgetWithText(ElevatedButton, 'Criar conta');
    await tester.tap(registerButton);
    await tester.pump();

    expect(find.text('Nome obrigatório'), findsOneWidget);
    expect(find.text('E-mail obrigatório'), findsOneWidget);
    expect(find.text('Senha obrigatória'), findsOneWidget);
  });

  testWidgets('RegisterScreen validates password confirmation mismatch', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: RegisterScreen(),
        ),
      ),
    );

    await tester.enterText(find.widgetWithText(TextFormField, 'Nome completo'), 'Alice');
    await tester.enterText(find.widgetWithText(TextFormField, 'E-mail'), 'alice@example.com');
    await tester.enterText(find.widgetWithText(TextFormField, 'Senha'), 'password123');
    await tester.enterText(find.widgetWithText(TextFormField, 'Confirmar senha'), 'different123');

    final registerButton = find.widgetWithText(ElevatedButton, 'Criar conta');
    await tester.tap(registerButton);
    await tester.pump();

    expect(find.text('As senhas não coincidem'), findsOneWidget);
  });
}
