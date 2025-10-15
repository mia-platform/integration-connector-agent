#!/bin/bash

# Script to update license headers from Apache 2.0 to AGPL + Commercial

# New license header
read -r -d '' NEW_HEADER << 'EOF'
// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
//
// This file is part of integration-connector-agent.
//
// integration-connector-agent is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// integration-connector-agent is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with integration-connector-agent. If not, see <https://www.gnu.org/licenses/>.
//
// Alternatively, this file may be used under the terms of a commercial license
// available from Mia-Platform. For inquiries, contact licensing@mia-platform.eu.
EOF

# Find all Go files and update their headers
find . -name "*.go" -type f | while read -r file; do
    echo "Updating $file"
    
    # Check if file has a license header
    if head -20 "$file" | grep -q "Apache\|Copyright"; then
        # Find where the license header ends (look for the first non-comment line after copyright)
        header_end=$(awk '
            /^\/\/ Copyright/ { in_header=1; next }
            in_header && /^\/\/ / { next }
            in_header && /^\/\*/ { in_multiline=1; next }
            in_multiline && /\*\// { in_multiline=0; next }
            in_multiline { next }
            in_header && /^$/ { next }
            in_header { print NR-1; exit }
            !in_header && /^package|^\/\/ \+build|^\/\/go:build/ { print NR-1; exit }
        ' "$file")
        
        if [ -n "$header_end" ] && [ "$header_end" -gt 0 ]; then
            # Create temp file with new header + rest of file
            {
                echo "$NEW_HEADER"
                echo ""
                tail -n +"$((header_end + 1))" "$file"
            } > "$file.tmp"
            mv "$file.tmp" "$file"
        else
            # No clear header found, just prepend new header
            {
                echo "$NEW_HEADER"
                echo ""
                cat "$file"
            } > "$file.tmp"
            mv "$file.tmp" "$file"
        fi
    else
        # No license header found, prepend new one
        {
            echo "$NEW_HEADER"
            echo ""
            cat "$file"
        } > "$file.tmp"
        mv "$file.tmp" "$file"
    fi
done

echo "License headers updated successfully!"