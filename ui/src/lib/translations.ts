import i18n from 'sveltekit-i18n';
import type Config from 'sveltekit-i18n';
const { MODE } = import.meta.env;

export const config: Config = {
    fallbackLocale: 'en',
    loaders: [
	{
	    locale: 'en',
	    key: '',
	    loader: async () => {
                if (MODE == 'development'){
                    return await (await fetch('/src/lib/locales/en.json')).json()
                } else {
                    return (await import('./locales/en.json')).default
                }
            }

	},
	{
	    locale: 'fr',
	    key: '',
	    loader: async () => (await import('./locales/fr.json')).default
	}
    ]
};

export const { t, locales, locale, loadTranslations } = new i18n(config);
