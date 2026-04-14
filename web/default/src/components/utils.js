import { copy } from '../helpers';

export async function copyToClipboard(text) {
  await copy(text);
}
