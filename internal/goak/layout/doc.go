// Package layout provides a two-pass layout system for nested containers.
//
// Sizing: each dimension (width/height) can be Static (fixed px), Percent (0–100 of parent),
// or Auto (share remaining space with other Auto siblings).
//
// Two passes (similar to Clay):
//   - Pass 1 (size): resolve each node's width and height from parent-available space.
//   - Pass 2 (position): assign x,y; children are stacked vertically.
//
// Usage: build a tree with NewContainer, then call Layout(root, viewW, viewH) on init
// and on every window resize. Each container's Bounds is filled with the computed Rect.
package layout
