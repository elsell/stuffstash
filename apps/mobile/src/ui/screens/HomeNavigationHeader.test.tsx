import React from 'react';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import { HomeNavigationHeader } from './HomeNavigationHeader';

it('orders Add, Notifications, Profile and gives the inventory selector padded, bounded space', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  try {
    await h.render(<HomeNavigationHeader dashboard={{ inventoryName: 'A long household inventory', tenantName: 'Home', canAdd: true }} notificationAction={{ kind: 'notifications', label: 'Notifications', onPress: () => {} }} />);
    expect(h.allByType('Pressable').map(node => node.props.accessibilityLabel)).toEqual(['Current inventory A long household inventory, tenant Home. Switch inventory', 'Add an asset', 'Notifications', 'Open account and settings']);
    const options = navigationOptions().at(-1) as { headerLeft: () => React.ReactElement };
    await h.render(options.headerLeft());
    const selector = h.byLabel('Current inventory A long household inventory, tenant Home. Switch inventory');
    const style = Object.assign({}, ...selector!.props.style);
    expect(style.paddingHorizontal).toBe(12);
    expect(style.minHeight).toBeGreaterThanOrEqual(44);
    expect(style.width).toBeLessThanOrEqual(160);
    expect(h.byText('A long household inventory')?.props.numberOfLines).toBe(1);
  } finally { await h.unmount(); }
});

it('settles navigation feedback while inventory context and notification commands change', async () => {
  const h = new MobileRenderHarness(); resetNavigation(); const calls: string[] = [];
  const render = (inventoryName: string, canAdd = true) => <NavigationOptionFeedback render={() => <HomeNavigationHeader
    dashboard={{ inventoryName, tenantName: 'Home', canAdd }}
    notificationAction={{ kind: 'notifications', label: 'Notifications', badgeCount: canAdd ? 1 : 2, onPress: () => calls.push(inventoryName) }} />} />;
  try {
    await h.render(render('Garage'));
    const notifications = h.byLabel('Notifications')!.props.onPress;
    await h.render(render('Kitchen', false));
    await h.run(notifications);
    expect(calls).toEqual(['Kitchen']);
    expect(h.byLabel('Current inventory Kitchen, tenant Home. Switch inventory')).toBeDefined();
    expect(h.byLabel('Add an asset')).toBeUndefined();
    expect(h.byLabel('Open account and settings')).toBeDefined();
  } finally { await h.unmount(); resetNavigation(); }
});
