import { grey } from '@ant-design/colors';
import { Typography as TypographyComponent } from 'antd';
import { themeColors } from 'constants/theme';
import styled from 'styled-components';

export const HeaderContainer = styled.div<{ hover: boolean }>`
	width: 100%;
	text-align: center;
	background: ${({ hover }): string => (hover ? `${grey[0]}66` : 'inherit')};
	padding: 0.25rem 0;
	font-size: 0.8rem;
	cursor: all-scroll;
	position: absolute;
	top: 0;
	left: 0;
`;

export const HeaderContentContainer = styled.span`
	cursor: pointer;
	position: relative;
	text-align: center;
`;

export const ArrowContainer = styled.span<{ hover: boolean }>`
	visibility: ${({ hover }): string => (hover ? 'visible' : 'hidden')};
	position: absolute;
	right: -1rem;
`;

export const ThesholdContainer = styled.span`
	margin-top: -0.3rem;
`;

export const DisplayThresholdContainer = styled.div`
	display: flex;
	align-items: center;
	width: auto;
	justify-content: space-between;
`;

export const WidgetHeaderContainer = styled.div`
	display: flex;
	flex-direction: row-reverse;
	align-items: center;
`;

export const Typography = styled(TypographyComponent)`
	&&& {
		color: ${themeColors.white};
		width: auto;
		margin-left: 0.2rem;
	}
`;

export const TypographHeading = styled(TypographyComponent)`
	&&& {
		color: ${grey[2]};
	}
`;
