import { StatusBar } from 'expo-status-bar';
import { SafeAreaView, StyleSheet, Text, View } from 'react-native';

export default function App() {
  return (
    <SafeAreaView style={styles.shell}>
      <StatusBar style="dark" />
      <View style={styles.content}>
        <Text style={styles.kicker}>Stuff Stash Mobile</Text>
        <Text style={styles.title}>Hello from Expo Go.</Text>
        <Text style={styles.body}>
          If you can see this on your iPhone, the local mobile development loop is alive.
        </Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: '#f7f5ef'
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    padding: 28
  },
  kicker: {
    color: '#396a63',
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    marginBottom: 12,
    textTransform: 'uppercase'
  },
  title: {
    color: '#1f2a27',
    fontSize: 34,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 40,
    marginBottom: 14
  },
  body: {
    color: '#4f5f59',
    fontSize: 17,
    lineHeight: 25,
    maxWidth: 320
  }
});
