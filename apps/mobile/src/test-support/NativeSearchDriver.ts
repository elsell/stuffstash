import { navigationOptions, subscribeNavigationOptions } from './navigation';

type SearchOptions = {
  ref: { current: unknown };
  onChangeText: (event: { nativeEvent: { text: string } }) => void;
  onClose: () => void;
  onFocus: () => void;
  placement: string;
};

/** Controlled native search field: observes header options and applies native commands. */
export class NativeSearchDriver {
  text = '';
  options: SearchOptions | undefined;
  private readonly unsubscribe = subscribeNavigationOptions(() => {
    const update = navigationOptions().at(-1) as { headerSearchBarOptions?: SearchOptions };
    if (!Object.hasOwn(update, 'headerSearchBarOptions')) return;
    if (this.options) this.options.ref.current = null;
    this.options = update.headerSearchBarOptions;
    if (this.options) this.options.ref.current = {
      setText: (text: string) => { this.text = text; },
      clearText: () => { this.text = ''; },
      blur: () => {}
    };
  });
  change(text: string) {
    if (!this.options) throw new Error('Native search is unavailable');
    this.options.onFocus();
    this.text = text;
    this.options.onChangeText({ nativeEvent: { text } });
  }
  dispose() {
    this.unsubscribe();
    if (this.options) this.options.ref.current = null;
  }
}
