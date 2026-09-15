import { Component, useEffect, useState, type ErrorInfo, type ReactNode } from 'react';
import { ScrollView, Text } from 'react-native';
import { router } from 'expo-router';
import { AddAssetScreen } from '../src/ui/screens/AddAssetScreen';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AddAssetContextQuery } from '../src/application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../src/application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../src/application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../src/application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../src/application/add/PhotoSelectionQuery';

export function AddAssetFixture() {
  const [fixture] = useState(() => {
    const context = { tenantId: 'audit-tenant', tenantName: 'Audit household', inventoryId: 'audit-inventory', inventoryName: 'Audit inventory', canAdd: true, assetTags: [] };
    return {
      context,
      client: createMobileQueryClient(),
      contextQuery: new AddAssetContextQuery({ getAddAssetContext: async () => context }),
      scopeQuery: new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'audit-principal' }) }),
      draftStore: new InMemoryAddAssetDraftStore('audit'),
      parents: new ParentLookupQuery({ listParentCandidates: async () => [] }),
      photos: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => fixture.context}>
    <AddFixtureErrorBoundary><AddAssetScreen inventoryAssetTypesQuery={{ execute: async () => [] }}
      addAssetContextQuery={fixture.contextQuery} addDraftScopeQuery={fixture.scopeQuery}
      addAssetDraftStore={fixture.draftStore} parentLookupQuery={fixture.parents} photoSelectionQuery={fixture.photos}
      createAssetCommand={{ execute: async input => {
        await new Promise(resolve => setTimeout(resolve, 5000));
        throw new Error(`Rejected draft: ${input.title}`);
      } }} onDismiss={() => router.back()} /></AddFixtureErrorBoundary>
  </MobileServerStateProvider>;
}


class AddFixtureErrorBoundary extends Component<{ children: ReactNode }, { message?: string; componentStack?: string }> {
  state: { message?: string; componentStack?: string } = {};
  static getDerivedStateFromError(error: Error) { return { message: error.message }; }
  componentDidCatch(_error: Error, info: ErrorInfo) { this.setState({ componentStack: info.componentStack ?? undefined }); }
  render() {
    if (this.state.message !== undefined) return <ScrollView>
      <Text accessibilityRole="header">Add fixture render failed</Text>
      <Text>{this.state.message}</Text><Text>{this.state.componentStack}</Text>
    </ScrollView>;
    return this.props.children;
  }
}
