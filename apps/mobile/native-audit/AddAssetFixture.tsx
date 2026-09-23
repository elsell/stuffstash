import { assetId } from '../src/domain/assets/AssetSummary';
import { returnToPreviousOrHome } from '../src/ui/navigation/returnToPreviousOrHome';
import { QueryReadinessDiagnostics } from './QueryReadinessDiagnostics';
import { Component, useEffect, useState, type ErrorInfo, type ReactNode } from 'react';
import { Image, ScrollView, Text } from 'react-native';
import { router } from 'expo-router';
import { AddAssetScreen } from '../src/ui/screens/AddAssetScreen';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AddAssetContextQuery } from '../src/application/add/AddAssetContextQuery';
import { AddDraftScopeQuery } from '../src/application/add/AddDraftScopeQuery';
import { InMemoryAddAssetDraftStore } from '../src/application/add/AddAssetDraftStore';
import { ParentLookupQuery } from '../src/application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../src/application/add/PhotoSelectionQuery';

export function AddDestinationFixture() { return <AddAssetFixture destinationAudit />; }

export function AddAssetFixture({ destinationAudit = false }: { readonly destinationAudit?: boolean } = {}) {
  const [fixture] = useState(() => {
    let selectedPhotos = false;
    let creations = 0;
    const destinations = Array.from({ length: 14 }, (_, i) => ({
      id: assetId(`destination-${i + 1}`), title: `Shelf ${i + 1}`, kind: 'location' as const,
      lifecycleState: 'active' as const, locationLabel: 'Garage', locationTrail: ['Inventory', 'Garage', `Shelf ${i + 1}`],
      parentLocationTrail: [], description: '', updatedAtLabel: '', hasPhoto: false
    }));
    const context = { tenantId: 'audit-tenant', tenantName: 'Audit household', inventoryId: 'audit-inventory', inventoryName: 'Audit inventory', canAdd: true, assetTags: Array.from({ length: 14 }, (_, index) => ({ id: `tag-${index + 1}`, key: `tag-${index + 1}`, displayName: `Tag ${index + 1}` })) };
    return {
      context,
      create: async (title: string) => {
        if (++creations === 1) throw new Error('Place creation unavailable. Try again.');
        return { id: 'created-destination', title, message: 'Place created' };
      },
      client: createMobileQueryClient(),
      contextQuery: new AddAssetContextQuery({ getAddAssetContext: async () => context }),
      scopeQuery: new AddDraftScopeQuery({ getCurrentPrincipal: async () => ({ id: 'audit-principal' }) }),
      draftStore: new InMemoryAddAssetDraftStore('audit'),
      parents: new ParentLookupQuery({ listParentCandidates: async query => destinationAudit ? destinations.filter(parent => parent.title.toLowerCase().includes(query.toLowerCase())) : [] }),
      photos: new PhotoSelectionQuery({
        selectFromLibrary: async () => {
          if (selectedPhotos) return [];
          selectedPhotos = true;
          const uri = Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri;
          return [1, 2].map(index => ({ id: `audit-draft-photo-${index}`, uri,
            fileName: `audit-draft-photo-${index}.png`, contentType: 'image/png' as const, sizeBytes: 1 }));
        },
        captureFromCamera: async () => []
      })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => fixture.context}>
    <AddFixtureErrorBoundary><AddAssetScreen initialParent={destinationAudit ? {
      id: 'garage', title: 'Garage', kind: 'location', subtitle: 'House', pathLabel: 'House / Garage', selectionHint: 'Location', willPromoteToContainer: false
    } : undefined} inventoryAssetTypesQuery={{ execute: async () => [] }}
      addAssetContextQuery={fixture.contextQuery} addDraftScopeQuery={fixture.scopeQuery}
      addAssetDraftStore={fixture.draftStore} parentLookupQuery={fixture.parents} photoSelectionQuery={fixture.photos}
      createAssetCommand={{ execute: async input => {
        if (destinationAudit && input.kind === 'location') return fixture.create(input.title);
        await new Promise(resolve => setTimeout(resolve, destinationAudit ? 50 : 5000));
        throw new Error(`Rejected draft: ${input.title}`);
      } }} onDismiss={() => returnToPreviousOrHome(router)} /></AddFixtureErrorBoundary>
    <QueryReadinessDiagnostics client={fixture.client} />
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
