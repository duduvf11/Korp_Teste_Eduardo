import { Routes } from '@angular/router';
import { NotasFiscais } from './notas-fiscais/notas-fiscais';
import { Produtos } from './produtos/produtos';

export const routes: Routes = [
	{
		path: '',
		pathMatch: 'full',
		redirectTo: 'produtos',
	},
	{
		path: 'produtos',
		component: Produtos,
	},
	{
		path: 'notas-fiscais',
		component: NotasFiscais,
	},
	{
		path: '**',
		redirectTo: 'produtos',
	},
];
